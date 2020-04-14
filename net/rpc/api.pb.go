// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api.proto

package rpc

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type AddrInfo struct {
	ID                   string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Addrs                []string `protobuf:"bytes,2,rep,name=addrs,proto3" json:"addrs,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AddrInfo) Reset()         { *m = AddrInfo{} }
func (m *AddrInfo) String() string { return proto.CompactTextString(m) }
func (*AddrInfo) ProtoMessage()    {}
func (*AddrInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{0}
}

func (m *AddrInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AddrInfo.Unmarshal(m, b)
}
func (m *AddrInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AddrInfo.Marshal(b, m, deterministic)
}
func (m *AddrInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AddrInfo.Merge(m, src)
}
func (m *AddrInfo) XXX_Size() int {
	return xxx_messageInfo_AddrInfo.Size(m)
}
func (m *AddrInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_AddrInfo.DiscardUnknown(m)
}

var xxx_messageInfo_AddrInfo proto.InternalMessageInfo

func (m *AddrInfo) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *AddrInfo) GetAddrs() []string {
	if m != nil {
		return m.Addrs
	}
	return nil
}

type ListenAddrRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListenAddrRequest) Reset()         { *m = ListenAddrRequest{} }
func (m *ListenAddrRequest) String() string { return proto.CompactTextString(m) }
func (*ListenAddrRequest) ProtoMessage()    {}
func (*ListenAddrRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{1}
}

func (m *ListenAddrRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListenAddrRequest.Unmarshal(m, b)
}
func (m *ListenAddrRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListenAddrRequest.Marshal(b, m, deterministic)
}
func (m *ListenAddrRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListenAddrRequest.Merge(m, src)
}
func (m *ListenAddrRequest) XXX_Size() int {
	return xxx_messageInfo_ListenAddrRequest.Size(m)
}
func (m *ListenAddrRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListenAddrRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListenAddrRequest proto.InternalMessageInfo

type ListenAddrReply struct {
	AddrInfo             *AddrInfo `protobuf:"bytes,1,opt,name=addrInfo,proto3" json:"addrInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ListenAddrReply) Reset()         { *m = ListenAddrReply{} }
func (m *ListenAddrReply) String() string { return proto.CompactTextString(m) }
func (*ListenAddrReply) ProtoMessage()    {}
func (*ListenAddrReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{2}
}

func (m *ListenAddrReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListenAddrReply.Unmarshal(m, b)
}
func (m *ListenAddrReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListenAddrReply.Marshal(b, m, deterministic)
}
func (m *ListenAddrReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListenAddrReply.Merge(m, src)
}
func (m *ListenAddrReply) XXX_Size() int {
	return xxx_messageInfo_ListenAddrReply.Size(m)
}
func (m *ListenAddrReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ListenAddrReply.DiscardUnknown(m)
}

var xxx_messageInfo_ListenAddrReply proto.InternalMessageInfo

func (m *ListenAddrReply) GetAddrInfo() *AddrInfo {
	if m != nil {
		return m.AddrInfo
	}
	return nil
}

type PeersRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PeersRequest) Reset()         { *m = PeersRequest{} }
func (m *PeersRequest) String() string { return proto.CompactTextString(m) }
func (*PeersRequest) ProtoMessage()    {}
func (*PeersRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{3}
}

func (m *PeersRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeersRequest.Unmarshal(m, b)
}
func (m *PeersRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeersRequest.Marshal(b, m, deterministic)
}
func (m *PeersRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeersRequest.Merge(m, src)
}
func (m *PeersRequest) XXX_Size() int {
	return xxx_messageInfo_PeersRequest.Size(m)
}
func (m *PeersRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PeersRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PeersRequest proto.InternalMessageInfo

type PeersReply struct {
	Peers                []*AddrInfo `protobuf:"bytes,1,rep,name=peers,proto3" json:"peers,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *PeersReply) Reset()         { *m = PeersReply{} }
func (m *PeersReply) String() string { return proto.CompactTextString(m) }
func (*PeersReply) ProtoMessage()    {}
func (*PeersReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{4}
}

func (m *PeersReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PeersReply.Unmarshal(m, b)
}
func (m *PeersReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PeersReply.Marshal(b, m, deterministic)
}
func (m *PeersReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PeersReply.Merge(m, src)
}
func (m *PeersReply) XXX_Size() int {
	return xxx_messageInfo_PeersReply.Size(m)
}
func (m *PeersReply) XXX_DiscardUnknown() {
	xxx_messageInfo_PeersReply.DiscardUnknown(m)
}

var xxx_messageInfo_PeersReply proto.InternalMessageInfo

func (m *PeersReply) GetPeers() []*AddrInfo {
	if m != nil {
		return m.Peers
	}
	return nil
}

type FindPeerRequest struct {
	PeerID               string   `protobuf:"bytes,1,opt,name=peerID,proto3" json:"peerID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *FindPeerRequest) Reset()         { *m = FindPeerRequest{} }
func (m *FindPeerRequest) String() string { return proto.CompactTextString(m) }
func (*FindPeerRequest) ProtoMessage()    {}
func (*FindPeerRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{5}
}

func (m *FindPeerRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindPeerRequest.Unmarshal(m, b)
}
func (m *FindPeerRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindPeerRequest.Marshal(b, m, deterministic)
}
func (m *FindPeerRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindPeerRequest.Merge(m, src)
}
func (m *FindPeerRequest) XXX_Size() int {
	return xxx_messageInfo_FindPeerRequest.Size(m)
}
func (m *FindPeerRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FindPeerRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FindPeerRequest proto.InternalMessageInfo

func (m *FindPeerRequest) GetPeerID() string {
	if m != nil {
		return m.PeerID
	}
	return ""
}

type FindPeerReply struct {
	PeerInfo             *AddrInfo `protobuf:"bytes,1,opt,name=peerInfo,proto3" json:"peerInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *FindPeerReply) Reset()         { *m = FindPeerReply{} }
func (m *FindPeerReply) String() string { return proto.CompactTextString(m) }
func (*FindPeerReply) ProtoMessage()    {}
func (*FindPeerReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{6}
}

func (m *FindPeerReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindPeerReply.Unmarshal(m, b)
}
func (m *FindPeerReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindPeerReply.Marshal(b, m, deterministic)
}
func (m *FindPeerReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindPeerReply.Merge(m, src)
}
func (m *FindPeerReply) XXX_Size() int {
	return xxx_messageInfo_FindPeerReply.Size(m)
}
func (m *FindPeerReply) XXX_DiscardUnknown() {
	xxx_messageInfo_FindPeerReply.DiscardUnknown(m)
}

var xxx_messageInfo_FindPeerReply proto.InternalMessageInfo

func (m *FindPeerReply) GetPeerInfo() *AddrInfo {
	if m != nil {
		return m.PeerInfo
	}
	return nil
}

type ConnectPeerRequest struct {
	PeerInfo             *AddrInfo `protobuf:"bytes,1,opt,name=peerInfo,proto3" json:"peerInfo,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *ConnectPeerRequest) Reset()         { *m = ConnectPeerRequest{} }
func (m *ConnectPeerRequest) String() string { return proto.CompactTextString(m) }
func (*ConnectPeerRequest) ProtoMessage()    {}
func (*ConnectPeerRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{7}
}

func (m *ConnectPeerRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConnectPeerRequest.Unmarshal(m, b)
}
func (m *ConnectPeerRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConnectPeerRequest.Marshal(b, m, deterministic)
}
func (m *ConnectPeerRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConnectPeerRequest.Merge(m, src)
}
func (m *ConnectPeerRequest) XXX_Size() int {
	return xxx_messageInfo_ConnectPeerRequest.Size(m)
}
func (m *ConnectPeerRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ConnectPeerRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ConnectPeerRequest proto.InternalMessageInfo

func (m *ConnectPeerRequest) GetPeerInfo() *AddrInfo {
	if m != nil {
		return m.PeerInfo
	}
	return nil
}

type ConnectPeerReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConnectPeerReply) Reset()         { *m = ConnectPeerReply{} }
func (m *ConnectPeerReply) String() string { return proto.CompactTextString(m) }
func (*ConnectPeerReply) ProtoMessage()    {}
func (*ConnectPeerReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{8}
}

func (m *ConnectPeerReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConnectPeerReply.Unmarshal(m, b)
}
func (m *ConnectPeerReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConnectPeerReply.Marshal(b, m, deterministic)
}
func (m *ConnectPeerReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConnectPeerReply.Merge(m, src)
}
func (m *ConnectPeerReply) XXX_Size() int {
	return xxx_messageInfo_ConnectPeerReply.Size(m)
}
func (m *ConnectPeerReply) XXX_DiscardUnknown() {
	xxx_messageInfo_ConnectPeerReply.DiscardUnknown(m)
}

var xxx_messageInfo_ConnectPeerReply proto.InternalMessageInfo

type DisconnectPeerRequest struct {
	PeerID               string   `protobuf:"bytes,1,opt,name=peerID,proto3" json:"peerID,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DisconnectPeerRequest) Reset()         { *m = DisconnectPeerRequest{} }
func (m *DisconnectPeerRequest) String() string { return proto.CompactTextString(m) }
func (*DisconnectPeerRequest) ProtoMessage()    {}
func (*DisconnectPeerRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{9}
}

func (m *DisconnectPeerRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DisconnectPeerRequest.Unmarshal(m, b)
}
func (m *DisconnectPeerRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DisconnectPeerRequest.Marshal(b, m, deterministic)
}
func (m *DisconnectPeerRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DisconnectPeerRequest.Merge(m, src)
}
func (m *DisconnectPeerRequest) XXX_Size() int {
	return xxx_messageInfo_DisconnectPeerRequest.Size(m)
}
func (m *DisconnectPeerRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DisconnectPeerRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DisconnectPeerRequest proto.InternalMessageInfo

func (m *DisconnectPeerRequest) GetPeerID() string {
	if m != nil {
		return m.PeerID
	}
	return ""
}

type DisconnectPeerReply struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DisconnectPeerReply) Reset()         { *m = DisconnectPeerReply{} }
func (m *DisconnectPeerReply) String() string { return proto.CompactTextString(m) }
func (*DisconnectPeerReply) ProtoMessage()    {}
func (*DisconnectPeerReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{10}
}

func (m *DisconnectPeerReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DisconnectPeerReply.Unmarshal(m, b)
}
func (m *DisconnectPeerReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DisconnectPeerReply.Marshal(b, m, deterministic)
}
func (m *DisconnectPeerReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DisconnectPeerReply.Merge(m, src)
}
func (m *DisconnectPeerReply) XXX_Size() int {
	return xxx_messageInfo_DisconnectPeerReply.Size(m)
}
func (m *DisconnectPeerReply) XXX_DiscardUnknown() {
	xxx_messageInfo_DisconnectPeerReply.DiscardUnknown(m)
}

var xxx_messageInfo_DisconnectPeerReply proto.InternalMessageInfo

func init() {
	proto.RegisterType((*AddrInfo)(nil), "rpc.AddrInfo")
	proto.RegisterType((*ListenAddrRequest)(nil), "rpc.ListenAddrRequest")
	proto.RegisterType((*ListenAddrReply)(nil), "rpc.ListenAddrReply")
	proto.RegisterType((*PeersRequest)(nil), "rpc.PeersRequest")
	proto.RegisterType((*PeersReply)(nil), "rpc.PeersReply")
	proto.RegisterType((*FindPeerRequest)(nil), "rpc.FindPeerRequest")
	proto.RegisterType((*FindPeerReply)(nil), "rpc.FindPeerReply")
	proto.RegisterType((*ConnectPeerRequest)(nil), "rpc.ConnectPeerRequest")
	proto.RegisterType((*ConnectPeerReply)(nil), "rpc.ConnectPeerReply")
	proto.RegisterType((*DisconnectPeerRequest)(nil), "rpc.DisconnectPeerRequest")
	proto.RegisterType((*DisconnectPeerReply)(nil), "rpc.DisconnectPeerReply")
}

func init() { proto.RegisterFile("api.proto", fileDescriptor_00212fb1f9d3bf1c) }

var fileDescriptor_00212fb1f9d3bf1c = []byte{
	// 387 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x93, 0xdf, 0x6e, 0xa2, 0x50,
	0x10, 0xc6, 0x17, 0x88, 0x46, 0xc7, 0x7f, 0xeb, 0xf8, 0x67, 0x09, 0xd9, 0x0b, 0x73, 0xf6, 0x46,
	0x2f, 0x96, 0xdd, 0xda, 0xa6, 0x17, 0x8d, 0x89, 0xd1, 0xda, 0xa6, 0x24, 0x4d, 0x43, 0x88, 0x2f,
	0x40, 0xe1, 0xb4, 0x21, 0x21, 0x70, 0x0a, 0xa7, 0x69, 0x7d, 0x9d, 0xbe, 0x43, 0xdf, 0xaf, 0xe1,
	0x00, 0x8a, 0xa2, 0x17, 0xbd, 0x9c, 0xf9, 0xe6, 0xfb, 0x31, 0xe7, 0x9b, 0x00, 0x75, 0x9b, 0x79,
	0x3a, 0x8b, 0x42, 0x1e, 0xa2, 0x12, 0x31, 0x87, 0xfc, 0x87, 0xda, 0xc2, 0x75, 0x23, 0x23, 0x78,
	0x0a, 0xb1, 0x0d, 0xb2, 0xb1, 0x52, 0xa5, 0x91, 0x34, 0xae, 0x5b, 0xb2, 0xb1, 0xc2, 0x3e, 0x54,
	0x6c, 0xd7, 0x8d, 0x62, 0x55, 0x1e, 0x29, 0xe3, 0xba, 0x95, 0x16, 0xa4, 0x07, 0xdd, 0x7b, 0x2f,
	0xe6, 0x34, 0x48, 0x7c, 0x16, 0x7d, 0x79, 0xa5, 0x31, 0x27, 0x33, 0xe8, 0x14, 0x9b, 0xcc, 0xdf,
	0xe0, 0x04, 0x6a, 0x76, 0x46, 0x16, 0xcc, 0xc6, 0xb4, 0xa5, 0x47, 0xcc, 0xd1, 0xf3, 0xcf, 0x59,
	0x5b, 0x99, 0xb4, 0xa1, 0x69, 0x52, 0x1a, 0xc5, 0x39, 0xed, 0x0c, 0x20, 0xab, 0x13, 0xd0, 0x1f,
	0xa8, 0xb0, 0xa4, 0x52, 0xa5, 0x91, 0x52, 0xa6, 0xa4, 0x1a, 0x99, 0x40, 0xe7, 0xd6, 0x0b, 0xdc,
	0xc4, 0x96, 0x51, 0x70, 0x08, 0xd5, 0x44, 0xdb, 0x3e, 0x29, 0xab, 0xc8, 0x15, 0xb4, 0x76, 0xa3,
	0xd9, 0xa6, 0x42, 0x3a, 0xbd, 0x69, 0x2e, 0x93, 0x39, 0xe0, 0x75, 0x18, 0x04, 0xd4, 0xe1, 0xc5,
	0x2f, 0x7d, 0x03, 0x80, 0xf0, 0x73, 0x0f, 0xc0, 0xfc, 0x0d, 0xf9, 0x07, 0x83, 0x95, 0x17, 0x3b,
	0x65, 0xee, 0xa9, 0x17, 0x0c, 0xa0, 0x77, 0x68, 0x60, 0xfe, 0x66, 0xfa, 0x29, 0x83, 0xb2, 0x30,
	0x0d, 0x9c, 0x01, 0xec, 0x8e, 0x81, 0x43, 0xb1, 0x4a, 0xe9, 0x64, 0x5a, 0xbf, 0xd4, 0x4f, 0x76,
	0xf9, 0x81, 0x7f, 0xa1, 0x22, 0xc2, 0xc7, 0xae, 0x18, 0x28, 0x1e, 0x46, 0xeb, 0x14, 0x5b, 0xe9,
	0xf8, 0x25, 0xd4, 0xf2, 0x34, 0x31, 0x45, 0x1e, 0xdc, 0x41, 0xc3, 0x83, 0x6e, 0xea, 0x9b, 0x43,
	0xa3, 0x10, 0x04, 0xfe, 0x12, 0x43, 0xe5, 0x6c, 0xb5, 0x41, 0x59, 0x48, 0x01, 0x77, 0xd0, 0xde,
	0x0f, 0x01, 0x35, 0x31, 0x7a, 0x34, 0x4a, 0x4d, 0x3d, 0xaa, 0x09, 0xd2, 0xf2, 0x02, 0x7e, 0x7b,
	0xa1, 0xce, 0xe9, 0x3b, 0xf7, 0x7c, 0xaa, 0xb3, 0xf0, 0x8d, 0x46, 0xcf, 0x36, 0xa7, 0x7a, 0x40,
	0x79, 0xe2, 0x5a, 0x36, 0xcd, 0xbc, 0xf5, 0x40, 0xb9, 0x29, 0x7d, 0xc8, 0xca, 0x7a, 0x7d, 0xf3,
	0x58, 0x15, 0x7f, 0xd1, 0xf9, 0x57, 0x00, 0x00, 0x00, 0xff, 0xff, 0xa3, 0xc3, 0x0c, 0x28, 0x52,
	0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// APIClient is the client API for API service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type APIClient interface {
	ListenAddr(ctx context.Context, in *ListenAddrRequest, opts ...grpc.CallOption) (*ListenAddrReply, error)
	Peers(ctx context.Context, in *PeersRequest, opts ...grpc.CallOption) (*PeersReply, error)
	FindPeer(ctx context.Context, in *FindPeerRequest, opts ...grpc.CallOption) (*FindPeerReply, error)
	ConnectPeer(ctx context.Context, in *ConnectPeerRequest, opts ...grpc.CallOption) (*ConnectPeerReply, error)
	DisconnectPeer(ctx context.Context, in *DisconnectPeerRequest, opts ...grpc.CallOption) (*DisconnectPeerReply, error)
}

type aPIClient struct {
	cc *grpc.ClientConn
}

func NewAPIClient(cc *grpc.ClientConn) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) ListenAddr(ctx context.Context, in *ListenAddrRequest, opts ...grpc.CallOption) (*ListenAddrReply, error) {
	out := new(ListenAddrReply)
	err := c.cc.Invoke(ctx, "/rpc.API/ListenAddr", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) Peers(ctx context.Context, in *PeersRequest, opts ...grpc.CallOption) (*PeersReply, error) {
	out := new(PeersReply)
	err := c.cc.Invoke(ctx, "/rpc.API/Peers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) FindPeer(ctx context.Context, in *FindPeerRequest, opts ...grpc.CallOption) (*FindPeerReply, error) {
	out := new(FindPeerReply)
	err := c.cc.Invoke(ctx, "/rpc.API/FindPeer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) ConnectPeer(ctx context.Context, in *ConnectPeerRequest, opts ...grpc.CallOption) (*ConnectPeerReply, error) {
	out := new(ConnectPeerReply)
	err := c.cc.Invoke(ctx, "/rpc.API/ConnectPeer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) DisconnectPeer(ctx context.Context, in *DisconnectPeerRequest, opts ...grpc.CallOption) (*DisconnectPeerReply, error) {
	out := new(DisconnectPeerReply)
	err := c.cc.Invoke(ctx, "/rpc.API/DisconnectPeer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// APIServer is the server API for API service.
type APIServer interface {
	ListenAddr(context.Context, *ListenAddrRequest) (*ListenAddrReply, error)
	Peers(context.Context, *PeersRequest) (*PeersReply, error)
	FindPeer(context.Context, *FindPeerRequest) (*FindPeerReply, error)
	ConnectPeer(context.Context, *ConnectPeerRequest) (*ConnectPeerReply, error)
	DisconnectPeer(context.Context, *DisconnectPeerRequest) (*DisconnectPeerReply, error)
}

// UnimplementedAPIServer can be embedded to have forward compatible implementations.
type UnimplementedAPIServer struct {
}

func (*UnimplementedAPIServer) ListenAddr(ctx context.Context, req *ListenAddrRequest) (*ListenAddrReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListenAddr not implemented")
}
func (*UnimplementedAPIServer) Peers(ctx context.Context, req *PeersRequest) (*PeersReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Peers not implemented")
}
func (*UnimplementedAPIServer) FindPeer(ctx context.Context, req *FindPeerRequest) (*FindPeerReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FindPeer not implemented")
}
func (*UnimplementedAPIServer) ConnectPeer(ctx context.Context, req *ConnectPeerRequest) (*ConnectPeerReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConnectPeer not implemented")
}
func (*UnimplementedAPIServer) DisconnectPeer(ctx context.Context, req *DisconnectPeerRequest) (*DisconnectPeerReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisconnectPeer not implemented")
}

func RegisterAPIServer(s *grpc.Server, srv APIServer) {
	s.RegisterService(&_API_serviceDesc, srv)
}

func _API_ListenAddr_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListenAddrRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).ListenAddr(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.API/ListenAddr",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).ListenAddr(ctx, req.(*ListenAddrRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_Peers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PeersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).Peers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.API/Peers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).Peers(ctx, req.(*PeersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_FindPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FindPeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).FindPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.API/FindPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).FindPeer(ctx, req.(*FindPeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_ConnectPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConnectPeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).ConnectPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.API/ConnectPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).ConnectPeer(ctx, req.(*ConnectPeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_DisconnectPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisconnectPeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).DisconnectPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/rpc.API/DisconnectPeer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).DisconnectPeer(ctx, req.(*DisconnectPeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _API_serviceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.API",
	HandlerType: (*APIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListenAddr",
			Handler:    _API_ListenAddr_Handler,
		},
		{
			MethodName: "Peers",
			Handler:    _API_Peers_Handler,
		},
		{
			MethodName: "FindPeer",
			Handler:    _API_FindPeer_Handler,
		},
		{
			MethodName: "ConnectPeer",
			Handler:    _API_ConnectPeer_Handler,
		},
		{
			MethodName: "DisconnectPeer",
			Handler:    _API_DisconnectPeer_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
