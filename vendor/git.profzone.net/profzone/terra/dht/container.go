package dht

import (
	"container/list"
	"sync"
)

type mapItem struct {
	key interface{}
	val interface{}
}

// SyncedMap represents a goroutine-safe map.
type SyncedMap struct {
	*sync.RWMutex
	data map[interface{}]interface{}
}

// NewSyncedMap returns a SyncedMap pointer.
func NewSyncedMap() *SyncedMap {
	return &SyncedMap{
		RWMutex: &sync.RWMutex{},
		data:    make(map[interface{}]interface{}),
	}
}

// Get returns the value mapped to key.
func (smap *SyncedMap) Get(key interface{}) (val interface{}, ok bool) {
	smap.RLock()
	defer smap.RUnlock()

	val, ok = smap.data[key]
	return
}

// Has returns whether the SyncedMap contains the key.
func (smap *SyncedMap) Has(key interface{}) bool {
	_, ok := smap.Get(key)
	return ok
}

// Set sets pair {key: val}.
func (smap *SyncedMap) Set(key interface{}, val interface{}) {
	smap.Lock()
	defer smap.Unlock()

	smap.data[key] = val
}

// Delete deletes the key in the map.
func (smap *SyncedMap) Delete(key interface{}) {
	smap.Lock()
	defer smap.Unlock()

	delete(smap.data, key)
}

// DeleteMulti deletes keys in batch.
func (smap *SyncedMap) DeleteMulti(keys []interface{}) {
	smap.Lock()
	defer smap.Unlock()

	for _, key := range keys {
		delete(smap.data, key)
	}
}

// Clear resets the data.
func (smap *SyncedMap) Clear() {
	smap.Lock()
	defer smap.Unlock()

	smap.data = make(map[interface{}]interface{})
}

// Iter returns a chan which output all items.
func (smap *SyncedMap) Iter() <-chan mapItem {
	ch := make(chan mapItem)
	go func() {
		smap.RLock()
		for key, val := range smap.data {
			ch <- mapItem{
				key: key,
				val: val,
			}
		}
		smap.RUnlock()
		close(ch)
	}()
	return ch
}

// Len returns the length of SyncedMap.
func (smap *SyncedMap) Len() int {
	smap.RLock()
	defer smap.RUnlock()

	return len(smap.data)
}

// SyncedList represents a goroutine-safe list.
type SyncedList struct {
	*sync.RWMutex
	queue *list.List
}

// NewSyncedList returns a SyncedList pointer.
func NewSyncedList() *SyncedList {
	return &SyncedList{
		RWMutex: &sync.RWMutex{},
		queue:   list.New(),
	}
}

// Front returns the first element of slist.
func (slist *SyncedList) Front() *list.Element {
	slist.RLock()
	defer slist.RUnlock()

	return slist.queue.Front()
}

// Back returns the last element of slist.
func (slist *SyncedList) Back() *list.Element {
	slist.RLock()
	defer slist.RUnlock()

	return slist.queue.Back()
}

// PushFront pushs an element to the head of slist.
func (slist *SyncedList) PushFront(v interface{}) *list.Element {
	slist.Lock()
	defer slist.Unlock()

	return slist.queue.PushFront(v)
}

// PushBack pushs an element to the tail of slist.
func (slist *SyncedList) PushBack(v interface{}) *list.Element {
	slist.Lock()
	defer slist.Unlock()

	return slist.queue.PushBack(v)
}

// InsertBefore inserts v before mark.
func (slist *SyncedList) InsertBefore(
	v interface{}, mark *list.Element) *list.Element {

	slist.Lock()
	defer slist.Unlock()

	return slist.queue.InsertBefore(v, mark)
}

// InsertAfter inserts v after mark.
func (slist *SyncedList) InsertAfter(
	v interface{}, mark *list.Element) *list.Element {

	slist.Lock()
	defer slist.Unlock()

	return slist.queue.InsertAfter(v, mark)
}

// Remove removes e from the slist.
func (slist *SyncedList) Remove(e *list.Element) interface{} {
	slist.Lock()
	defer slist.Unlock()

	return slist.queue.Remove(e)
}

// Clear resets the list queue.
func (slist *SyncedList) Clear() {
	slist.Lock()
	defer slist.Unlock()

	slist.queue.Init()
}

// Len returns length of the slist.
func (slist *SyncedList) Len() int {
	slist.RLock()
	defer slist.RUnlock()

	return slist.queue.Len()
}

// Iter returns a chan which output all elements.
func (slist *SyncedList) Iter() <-chan *list.Element {
	ch := make(chan *list.Element)
	go func() {
		slist.RLock()
		for e := slist.queue.Front(); e != nil; e = e.Next() {
			ch <- e
		}
		slist.RUnlock()
		close(ch)
	}()
	return ch
}

// KeyedDeque represents a keyed deque.
type KeyedDeque struct {
	*sync.RWMutex
	*SyncedList
	index         map[interface{}]*list.Element
	invertedIndex map[*list.Element]interface{}
}

// NewKeyedDeque returns a NewKeyedDeque pointer.
func NewKeyedDeque() *KeyedDeque {
	return &KeyedDeque{
		RWMutex:       &sync.RWMutex{},
		SyncedList:    NewSyncedList(),
		index:         make(map[interface{}]*list.Element),
		invertedIndex: make(map[*list.Element]interface{}),
	}
}

// Push pushs a keyed-value to the end of deque.
func (deque *KeyedDeque) Push(key interface{}, val interface{}) {
	deque.Lock()
	defer deque.Unlock()

	if e, ok := deque.index[key]; ok {
		deque.SyncedList.Remove(e)
	}
	deque.index[key] = deque.SyncedList.PushBack(val)
	deque.invertedIndex[deque.index[key]] = key
}

// Get returns the keyed value.
func (deque *KeyedDeque) Get(key interface{}) (*list.Element, bool) {
	deque.RLock()
	defer deque.RUnlock()

	v, ok := deque.index[key]
	return v, ok
}

// Has returns whether key already exists.
func (deque *KeyedDeque) HasKey(key interface{}) bool {
	_, ok := deque.Get(key)
	return ok
}

// Delete deletes a value named key.
func (deque *KeyedDeque) Delete(key interface{}) (v interface{}) {
	deque.RLock()
	e, ok := deque.index[key]
	deque.RUnlock()

	deque.Lock()
	defer deque.Unlock()

	if ok {
		v = deque.SyncedList.Remove(e)
		delete(deque.index, key)
		delete(deque.invertedIndex, e)
	}

	return
}

// Removes overwrites list.List.Remove.
func (deque *KeyedDeque) Remove(e *list.Element) (v interface{}) {
	deque.RLock()
	key, ok := deque.invertedIndex[e]
	deque.RUnlock()

	if ok {
		v = deque.Delete(key)
	}

	return
}

// Clear resets the deque.
func (deque *KeyedDeque) Clear() {
	deque.Lock()
	defer deque.Unlock()

	deque.SyncedList.Clear()
	deque.index = make(map[interface{}]*list.Element)
	deque.invertedIndex = make(map[*list.Element]interface{})
}
