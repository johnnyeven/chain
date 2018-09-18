package dht

import (
	"sync"
	"time"
	"strings"
	"container/heap"
	"git.profzone.net/profzone/terra/dht/util"
)

const maxPrefixLength = 160

type bucket struct {
	sync.RWMutex
	nodes          *KeyedDeque
	candidates     *KeyedDeque
	prefix         *Identity
	lastChangeTime time.Time
}

func newBucket(prefix *Identity) *bucket {
	return &bucket{
		nodes:          NewKeyedDeque(),
		candidates:     NewKeyedDeque(),
		prefix:         prefix,
		lastChangeTime: time.Now(),
	}
}

func (b *bucket) LastChangeTime() time.Time {
	b.RLock()
	defer b.RUnlock()

	return b.lastChangeTime
}

func (b *bucket) RandomChildID() string {
	prefixLen := b.prefix.Size / 8

	return strings.Join([]string{
		b.prefix.RawString()[:prefixLen],
		util.RandomString(20 - prefixLen),
	}, "")
}

func (b *bucket) UpdateTimestamp() {
	b.Lock()
	defer b.Unlock()

	b.lastChangeTime = time.Now()
}

func (b *bucket) Insert(node *Node) bool {
	isNew := !b.nodes.HasKey(node.ID.RawString())

	b.nodes.Push(node.ID.RawString(), node)
	b.UpdateTimestamp()

	return isNew
}

func (b *bucket) Replace(node *Node) {
	b.nodes.Delete(node.ID.RawString())
	b.UpdateTimestamp()

	if b.candidates.Len() == 0 {
		return
	}

	candidateNode := b.candidates.Remove(b.candidates.Back()).(*Node)

	insertedMiddle := false
	for e := range b.nodes.Iter() {
		if e.Value.(*Node).LastActiveTime.After(node.LastActiveTime) && !insertedMiddle {
			b.nodes.InsertBefore(candidateNode, e)
			insertedMiddle = true
		}
	}

	if !insertedMiddle {
		b.nodes.PushBack(candidateNode)
	}
}

func (b *bucket) Fresh(table *DistributedHashTable) {
	for e := range b.nodes.Iter() {
		node := e.Value.(*Node)
		if time.Since(node.LastActiveTime) > table.NodeExpiredAfter {
			table.PingFunc(node, table.GetTransport())
		}
	}
}

type routingTableNode struct {
	sync.RWMutex
	children []*routingTableNode
	bucket   *bucket
}

func newRoutingTableNode(prefix *Identity) *routingTableNode {
	return &routingTableNode{
		children: make([]*routingTableNode, 2),
		bucket:   newBucket(prefix),
	}
}

func (tableNode *routingTableNode) Child(index int) *routingTableNode {
	if index > 1 {
		return nil
	}

	tableNode.RLock()
	defer tableNode.RUnlock()

	return tableNode.children[index]
}

func (tableNode *routingTableNode) SetChild(index int, child *routingTableNode) {
	if index > 1 {
		return
	}

	tableNode.Lock()
	defer tableNode.Unlock()

	tableNode.children[index] = child
}

func (tableNode *routingTableNode) Bucket() *bucket {
	tableNode.RLock()
	defer tableNode.RUnlock()

	return tableNode.bucket
}

func (tableNode *routingTableNode) SetBucket(b *bucket) {
	tableNode.Lock()
	defer tableNode.Unlock()

	tableNode.bucket = b
}

func (tableNode *routingTableNode) Split() {
	prefixLen := tableNode.bucket.prefix.Size
	if prefixLen == maxPrefixLength {
		return
	}

	for i := 0; i < 2; i++ {
		tableNode.SetChild(i, newRoutingTableNode(NewIdentityCopy(tableNode.bucket.prefix, prefixLen+1)))
	}

	tableNode.Lock()
	tableNode.children[1].bucket.prefix.Set(prefixLen)
	tableNode.Unlock()

	for e := range tableNode.bucket.nodes.Iter() {
		node := e.Value.(*Node)
		tableNode.Child(node.ID.Bit(prefixLen)).bucket.nodes.PushBack(node)
	}

	for e := range tableNode.bucket.candidates.Iter() {
		node := e.Value.(*Node)
		tableNode.Child(node.ID.Bit(prefixLen)).bucket.candidates.PushBack(node)
	}

	for i := 0; i < 2; i++ {
		tableNode.Child(i).bucket.UpdateTimestamp()
	}
}

type routingTable struct {
	sync.RWMutex
	k             int
	root          *routingTableNode
	cachedNodes   *SyncedMap
	cachedBuckets *KeyedDeque
	table         *DistributedHashTable
	clearQueue    *SyncedList
}

func newRoutingTable(k int, table *DistributedHashTable) *routingTable {
	root := newRoutingTableNode(newIdentity(0))

	rt := &routingTable{
		k:             k,
		root:          root,
		cachedNodes:   NewSyncedMap(),
		cachedBuckets: NewKeyedDeque(),
		table:         table,
		clearQueue:    NewSyncedList(),
	}
	rt.cachedBuckets.Push(root.bucket.prefix.String(), root.bucket)
	return rt
}

func (rt *routingTable) Insert(node *Node) bool {
	rt.Lock()
	defer rt.Unlock()

	if rt.cachedNodes.Len() >= rt.table.MaxNodes {
		return false
	}

	var (
		next   *routingTableNode
		bucket *bucket
	)
	root := rt.root

	for prefixLen := 1; prefixLen <= maxPrefixLength; prefixLen++ {
		next = root.Child(node.ID.Bit(prefixLen - 1))

		if next != nil {
			root = next
		} else if root.bucket.nodes.Len() < rt.k ||
			root.bucket.nodes.HasKey(node.ID.RawString()) {

			isNew := root.bucket.Insert(node)

			rt.cachedNodes.Set(node.Addr.String(), node)
			rt.cachedBuckets.Push(root.bucket.prefix.String(), root.bucket)

			return isNew
		} else if root.bucket.prefix.Compare(node.ID, prefixLen-1) == 0 {
			root.Split()

			rt.cachedBuckets.Delete(root.bucket.prefix.String())
			root.SetBucket(nil)

			for i := 0; i < 2; i++ {
				bucket = root.Child(i).bucket
				rt.cachedBuckets.Push(bucket.prefix.String(), bucket)
			}

			root = root.Child(node.ID.Bit(prefixLen - 1))
		} else {
			root.bucket.candidates.PushBack(node)
			if root.bucket.candidates.Len() > rt.k {
				root.bucket.candidates.Remove(root.bucket.candidates.Front())
			}

			go root.bucket.Fresh(rt.table)

			return false
		}
	}

	return false
}

func (rt *routingTable) GetNeighbors(id *Identity, size int) []*Node {
	rt.RLock()
	nodes := make([]interface{}, 0, rt.cachedNodes.Len())
	for item := range rt.cachedNodes.Iter() {
		nodes = append(nodes, item.val.(*Node))
	}
	rt.RUnlock()

	neighbors := getTopK(nodes, id, size)

	result := make([]*Node, len(neighbors))
	for i, node := range neighbors {
		result[i] = node.(*Node)
	}
	return result
}

func (rt *routingTable) GetNeighborCompactInfos(id *Identity, size int) []string {
	neighbors := rt.GetNeighbors(id, size)
	infos := make([]string, len(neighbors))

	for i, node := range neighbors {
		infos[i] = node.CompactNodeInfo()
	}

	return infos
}

func (rt *routingTable) GetNodeBucketByID(id *Identity) (node *Node, bucket *bucket) {
	rt.RLock()
	defer rt.RUnlock()

	var next *routingTableNode
	root := rt.root

	for prefixLen := 1; prefixLen <= maxPrefixLength; prefixLen++ {
		next = root.Child(id.Bit(prefixLen - 1))
		if next != nil {
			root = next
			continue
		}
		v, ok := root.bucket.nodes.Get(id.RawString())
		if !ok {
			return
		}
		node, bucket = v.Value.(*Node), root.bucket
		return
	}

	return
}

func (rt *routingTable) GetNodeByAddress(address string) (node *Node, ok bool) {
	rt.RLock()
	defer rt.RUnlock()

	v, ok := rt.cachedNodes.Get(address)
	if ok {
		node = v.(*Node)
	}

	return
}

func (rt *routingTable) Remove(id *Identity) {
	if node, bucket := rt.GetNodeBucketByID(id); node != nil {
		bucket.Replace(node)
		rt.cachedNodes.Delete(node.Addr.String())
		rt.cachedBuckets.Push(bucket.prefix.String(), bucket)
	}
}

func (rt *routingTable) RemoveByAddr(address string) {
	v, ok := rt.cachedNodes.Get(address)
	if ok {
		rt.Remove(v.(*Node).ID)
	}
}

func (rt *routingTable) Fresh() {
	now := time.Now()

	for e := range rt.cachedBuckets.Iter() {
		bucket := e.Value.(*bucket)
		if now.Sub(bucket.LastChangeTime()) < rt.table.BucketExpiredAfter || bucket.nodes.Len() == 0 {
			continue
		}

		i := 0
		for e := range bucket.nodes.Iter() {
			if i < rt.table.RefreshNodeCount {
				node := e.Value.(*Node)
				rt.table.HandshakeFunc(node, rt.table.GetTransport(), []byte(bucket.RandomChildID()))
				rt.clearQueue.PushBack(node)
			}
			i++
		}
	}

	rt.clearQueue.Clear()
}

func (rt *routingTable) Len() int {
	rt.RLock()
	defer rt.RUnlock()

	return rt.cachedNodes.Len()
}

// Implementation of heap with heap.Interface.
type heapItem struct {
	distance *Identity
	value    interface{}
}

type topKHeap []*heapItem

func (kHeap topKHeap) Len() int {
	return len(kHeap)
}

func (kHeap topKHeap) Less(i, j int) bool {
	return kHeap[i].distance.Compare(kHeap[j].distance, maxPrefixLength) == 1
}

func (kHeap topKHeap) Swap(i, j int) {
	kHeap[i], kHeap[j] = kHeap[j], kHeap[i]
}

func (kHeap *topKHeap) Push(x interface{}) {
	*kHeap = append(*kHeap, x.(*heapItem))
}

func (kHeap *topKHeap) Pop() interface{} {
	n := len(*kHeap)
	x := (*kHeap)[n-1]
	*kHeap = (*kHeap)[:n-1]
	return x
}

// getTopK solves the top-k problem with heap. It's time complexity is
// O(n*log(k)). When n is large, time complexity will be too high, need to be
// optimized.
func getTopK(queue []interface{}, id *Identity, k int) []interface{} {
	topkHeap := make(topKHeap, 0, k+1)

	for _, value := range queue {
		node := value.(*Node)
		item := &heapItem{
			id.Xor(node.ID),
			value,
		}
		heap.Push(&topkHeap, item)
		if topkHeap.Len() > k {
			heap.Pop(&topkHeap)
		}
	}

	tops := make([]interface{}, topkHeap.Len())
	for i := len(tops) - 1; i >= 0; i-- {
		tops[i] = heap.Pop(&topkHeap).(*heapItem).value
	}

	return tops
}
