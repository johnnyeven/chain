package global

//go:generate libtools gen enum MessageType
// swagger:enum
type MessageType uint32

// 消息类型
const (
	MESSAGE_TYPE_UNKNOWN                      MessageType = iota
	MESSAGE_TYPE__HEARTBEAT                    // 心跳检测
	MESSAGE_TYPE__HEARTBEAT_ACK                // 心跳检测确认
	MESSAGE_TYPE__HELLO                        // 首次握手
	MESSAGE_TYPE__HELLO_ACK                    // 首次握手确认
	MESSAGE_TYPE__HELLO_TCP                    // TCP握手
	MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS  // 新节点接入广播
	MESSAGE_TYPE__NEW_PEERS_TO_PEER            // 节点介绍

	MESSAGE_TYPE__REQUEST_HEIGHT   // 获取最新节点高度
	MESSAGE_TYPE__BLOCKS_HASH      // 发送区块哈希
	MESSAGE_TYPE__GET_BLOCK        // 获取区块
	MESSAGE_TYPE__GET_BLOCK_ACK    // 响应获取区块
	MESSAGE_TYPE__NEW_TRANSACTION  // 新交易
)
