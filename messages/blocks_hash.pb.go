// Code generated by protoc-gen-go. DO NOT EDIT.
// source: blocks_hash.proto

package messages

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BlocksHash struct {
	Hashes               [][]byte `protobuf:"bytes,1,rep,name=hashes,proto3" json:"hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlocksHash) Reset()         { *m = BlocksHash{} }
func (m *BlocksHash) String() string { return proto.CompactTextString(m) }
func (*BlocksHash) ProtoMessage()    {}
func (*BlocksHash) Descriptor() ([]byte, []int) {
	return fileDescriptor_blocks_hash_78ab61e1d376201c, []int{0}
}
func (m *BlocksHash) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlocksHash.Unmarshal(m, b)
}
func (m *BlocksHash) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlocksHash.Marshal(b, m, deterministic)
}
func (dst *BlocksHash) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlocksHash.Merge(dst, src)
}
func (m *BlocksHash) XXX_Size() int {
	return xxx_messageInfo_BlocksHash.Size(m)
}
func (m *BlocksHash) XXX_DiscardUnknown() {
	xxx_messageInfo_BlocksHash.DiscardUnknown(m)
}

var xxx_messageInfo_BlocksHash proto.InternalMessageInfo

func (m *BlocksHash) GetHashes() [][]byte {
	if m != nil {
		return m.Hashes
	}
	return nil
}

func init() {
	proto.RegisterType((*BlocksHash)(nil), "messages.BlocksHash")
}

func init() { proto.RegisterFile("blocks_hash.proto", fileDescriptor_blocks_hash_78ab61e1d376201c) }

var fileDescriptor_blocks_hash_78ab61e1d376201c = []byte{
	// 86 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0x4c, 0xca, 0xc9, 0x4f,
	0xce, 0x2e, 0x8e, 0xcf, 0x48, 0x2c, 0xce, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0xc8,
	0x4d, 0x2d, 0x2e, 0x4e, 0x4c, 0x4f, 0x2d, 0x56, 0x52, 0xe1, 0xe2, 0x72, 0x02, 0x4b, 0x7b, 0x24,
	0x16, 0x67, 0x08, 0x89, 0x71, 0xb1, 0x81, 0x54, 0xa5, 0x16, 0x4b, 0x30, 0x2a, 0x30, 0x6b, 0xf0,
	0x04, 0x41, 0x79, 0x49, 0x6c, 0x60, 0x6d, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0x17, 0x81,
	0x7d, 0xde, 0x4b, 0x00, 0x00, 0x00,
}
