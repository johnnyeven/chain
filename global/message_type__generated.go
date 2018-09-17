package global

import (
	"bytes"
	"encoding"
	"errors"

	git_profzone_net_profzone_libtools_courier_enumeration "git.profzone.net/profzone/libtools/courier/enumeration"
)

var InvalidMessageType = errors.New("invalid MessageType")

func init() {
	git_profzone_net_profzone_libtools_courier_enumeration.RegisterEnums("MessageType", map[string]string{
		"BLOCKS_HASH":                 "发送区块哈希",
		"BROADCAST_NEW_PEER_TO_PEERS": "新节点接入广播",
		"GET_BLOCK":                   "获取区块",
		"GET_BLOCK_ACK":               "响应获取区块",
		"HEARTBEAT":                   "心跳检测",
		"HEARTBEAT_ACK":               "心跳检测确认",
		"HELLO":                       "首次握手",
		"HELLO_ACK":                   "首次握手确认",
		"HELLO_TCP":                   "TCP握手",
		"NEW_PEERS_TO_PEER":           "节点介绍",
		"NEW_TRANSACTION":             "新交易",
		"REQUEST_HEIGHT":              "获取最新节点高度",
	})
}

func ParseMessageTypeFromString(s string) (MessageType, error) {
	switch s {
	case "":
		return MESSAGE_TYPE_UNKNOWN, nil
	case "BLOCKS_HASH":
		return MESSAGE_TYPE__BLOCKS_HASH, nil
	case "BROADCAST_NEW_PEER_TO_PEERS":
		return MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS, nil
	case "GET_BLOCK":
		return MESSAGE_TYPE__GET_BLOCK, nil
	case "GET_BLOCK_ACK":
		return MESSAGE_TYPE__GET_BLOCK_ACK, nil
	case "HEARTBEAT":
		return MESSAGE_TYPE__HEARTBEAT, nil
	case "HEARTBEAT_ACK":
		return MESSAGE_TYPE__HEARTBEAT_ACK, nil
	case "HELLO":
		return MESSAGE_TYPE__HELLO, nil
	case "HELLO_ACK":
		return MESSAGE_TYPE__HELLO_ACK, nil
	case "HELLO_TCP":
		return MESSAGE_TYPE__HELLO_TCP, nil
	case "NEW_PEERS_TO_PEER":
		return MESSAGE_TYPE__NEW_PEERS_TO_PEER, nil
	case "NEW_TRANSACTION":
		return MESSAGE_TYPE__NEW_TRANSACTION, nil
	case "REQUEST_HEIGHT":
		return MESSAGE_TYPE__REQUEST_HEIGHT, nil
	}
	return MESSAGE_TYPE_UNKNOWN, InvalidMessageType
}

func ParseMessageTypeFromLabelString(s string) (MessageType, error) {
	switch s {
	case "":
		return MESSAGE_TYPE_UNKNOWN, nil
	case "发送区块哈希":
		return MESSAGE_TYPE__BLOCKS_HASH, nil
	case "新节点接入广播":
		return MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS, nil
	case "获取区块":
		return MESSAGE_TYPE__GET_BLOCK, nil
	case "响应获取区块":
		return MESSAGE_TYPE__GET_BLOCK_ACK, nil
	case "心跳检测":
		return MESSAGE_TYPE__HEARTBEAT, nil
	case "心跳检测确认":
		return MESSAGE_TYPE__HEARTBEAT_ACK, nil
	case "首次握手":
		return MESSAGE_TYPE__HELLO, nil
	case "首次握手确认":
		return MESSAGE_TYPE__HELLO_ACK, nil
	case "TCP握手":
		return MESSAGE_TYPE__HELLO_TCP, nil
	case "节点介绍":
		return MESSAGE_TYPE__NEW_PEERS_TO_PEER, nil
	case "新交易":
		return MESSAGE_TYPE__NEW_TRANSACTION, nil
	case "获取最新节点高度":
		return MESSAGE_TYPE__REQUEST_HEIGHT, nil
	}
	return MESSAGE_TYPE_UNKNOWN, InvalidMessageType
}

func (MessageType) EnumType() string {
	return "MessageType"
}

func (MessageType) Enums() map[int][]string {
	return map[int][]string{
		int(MESSAGE_TYPE__BLOCKS_HASH):                 {"BLOCKS_HASH", "发送区块哈希"},
		int(MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS): {"BROADCAST_NEW_PEER_TO_PEERS", "新节点接入广播"},
		int(MESSAGE_TYPE__GET_BLOCK):                   {"GET_BLOCK", "获取区块"},
		int(MESSAGE_TYPE__GET_BLOCK_ACK):               {"GET_BLOCK_ACK", "响应获取区块"},
		int(MESSAGE_TYPE__HEARTBEAT):                   {"HEARTBEAT", "心跳检测"},
		int(MESSAGE_TYPE__HEARTBEAT_ACK):               {"HEARTBEAT_ACK", "心跳检测确认"},
		int(MESSAGE_TYPE__HELLO):                       {"HELLO", "首次握手"},
		int(MESSAGE_TYPE__HELLO_ACK):                   {"HELLO_ACK", "首次握手确认"},
		int(MESSAGE_TYPE__HELLO_TCP):                   {"HELLO_TCP", "TCP握手"},
		int(MESSAGE_TYPE__NEW_PEERS_TO_PEER):           {"NEW_PEERS_TO_PEER", "节点介绍"},
		int(MESSAGE_TYPE__NEW_TRANSACTION):             {"NEW_TRANSACTION", "新交易"},
		int(MESSAGE_TYPE__REQUEST_HEIGHT):              {"REQUEST_HEIGHT", "获取最新节点高度"},
	}
}
func (v MessageType) String() string {
	switch v {
	case MESSAGE_TYPE_UNKNOWN:
		return ""
	case MESSAGE_TYPE__BLOCKS_HASH:
		return "BLOCKS_HASH"
	case MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS:
		return "BROADCAST_NEW_PEER_TO_PEERS"
	case MESSAGE_TYPE__GET_BLOCK:
		return "GET_BLOCK"
	case MESSAGE_TYPE__GET_BLOCK_ACK:
		return "GET_BLOCK_ACK"
	case MESSAGE_TYPE__HEARTBEAT:
		return "HEARTBEAT"
	case MESSAGE_TYPE__HEARTBEAT_ACK:
		return "HEARTBEAT_ACK"
	case MESSAGE_TYPE__HELLO:
		return "HELLO"
	case MESSAGE_TYPE__HELLO_ACK:
		return "HELLO_ACK"
	case MESSAGE_TYPE__HELLO_TCP:
		return "HELLO_TCP"
	case MESSAGE_TYPE__NEW_PEERS_TO_PEER:
		return "NEW_PEERS_TO_PEER"
	case MESSAGE_TYPE__NEW_TRANSACTION:
		return "NEW_TRANSACTION"
	case MESSAGE_TYPE__REQUEST_HEIGHT:
		return "REQUEST_HEIGHT"
	}
	return "UNKNOWN"
}

func (v MessageType) Label() string {
	switch v {
	case MESSAGE_TYPE_UNKNOWN:
		return ""
	case MESSAGE_TYPE__BLOCKS_HASH:
		return "发送区块哈希"
	case MESSAGE_TYPE__BROADCAST_NEW_PEER_TO_PEERS:
		return "新节点接入广播"
	case MESSAGE_TYPE__GET_BLOCK:
		return "获取区块"
	case MESSAGE_TYPE__GET_BLOCK_ACK:
		return "响应获取区块"
	case MESSAGE_TYPE__HEARTBEAT:
		return "心跳检测"
	case MESSAGE_TYPE__HEARTBEAT_ACK:
		return "心跳检测确认"
	case MESSAGE_TYPE__HELLO:
		return "首次握手"
	case MESSAGE_TYPE__HELLO_ACK:
		return "首次握手确认"
	case MESSAGE_TYPE__HELLO_TCP:
		return "TCP握手"
	case MESSAGE_TYPE__NEW_PEERS_TO_PEER:
		return "节点介绍"
	case MESSAGE_TYPE__NEW_TRANSACTION:
		return "新交易"
	case MESSAGE_TYPE__REQUEST_HEIGHT:
		return "获取最新节点高度"
	}
	return "UNKNOWN"
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*MessageType)(nil)

func (v MessageType) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, InvalidMessageType
	}
	return []byte(str), nil
}

func (v *MessageType) UnmarshalText(data []byte) (err error) {
	*v, err = ParseMessageTypeFromString(string(bytes.ToUpper(data)))
	return
}
