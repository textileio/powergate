// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: reputation/rpc/rpc.proto

package rpc

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type MinerScore struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Addr  string `protobuf:"bytes,1,opt,name=addr,proto3" json:"addr,omitempty"`
	Score int32  `protobuf:"varint,2,opt,name=score,proto3" json:"score,omitempty"`
}

func (x *MinerScore) Reset() {
	*x = MinerScore{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MinerScore) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MinerScore) ProtoMessage() {}

func (x *MinerScore) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MinerScore.ProtoReflect.Descriptor instead.
func (*MinerScore) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *MinerScore) GetAddr() string {
	if x != nil {
		return x.Addr
	}
	return ""
}

func (x *MinerScore) GetScore() int32 {
	if x != nil {
		return x.Score
	}
	return 0
}

type Index struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tipsetkey string              `protobuf:"bytes,1,opt,name=tipsetkey,proto3" json:"tipsetkey,omitempty"`
	Miners    map[string]*Slashes `protobuf:"bytes,2,rep,name=miners,proto3" json:"miners,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Index) Reset() {
	*x = Index{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index) ProtoMessage() {}

func (x *Index) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Index.ProtoReflect.Descriptor instead.
func (*Index) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *Index) GetTipsetkey() string {
	if x != nil {
		return x.Tipsetkey
	}
	return ""
}

func (x *Index) GetMiners() map[string]*Slashes {
	if x != nil {
		return x.Miners
	}
	return nil
}

type Slashes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Epochs []uint64 `protobuf:"varint,1,rep,packed,name=epochs,proto3" json:"epochs,omitempty"`
}

func (x *Slashes) Reset() {
	*x = Slashes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Slashes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Slashes) ProtoMessage() {}

func (x *Slashes) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Slashes.ProtoReflect.Descriptor instead.
func (*Slashes) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{2}
}

func (x *Slashes) GetEpochs() []uint64 {
	if x != nil {
		return x.Epochs
	}
	return nil
}

type AddSourceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id    string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Maddr string `protobuf:"bytes,2,opt,name=maddr,proto3" json:"maddr,omitempty"`
}

func (x *AddSourceRequest) Reset() {
	*x = AddSourceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddSourceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddSourceRequest) ProtoMessage() {}

func (x *AddSourceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddSourceRequest.ProtoReflect.Descriptor instead.
func (*AddSourceRequest) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{3}
}

func (x *AddSourceRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *AddSourceRequest) GetMaddr() string {
	if x != nil {
		return x.Maddr
	}
	return ""
}

type AddSourceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *AddSourceResponse) Reset() {
	*x = AddSourceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddSourceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddSourceResponse) ProtoMessage() {}

func (x *AddSourceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddSourceResponse.ProtoReflect.Descriptor instead.
func (*AddSourceResponse) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{4}
}

type GetTopMinersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Limit int32 `protobuf:"varint,1,opt,name=limit,proto3" json:"limit,omitempty"`
}

func (x *GetTopMinersRequest) Reset() {
	*x = GetTopMinersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTopMinersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTopMinersRequest) ProtoMessage() {}

func (x *GetTopMinersRequest) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTopMinersRequest.ProtoReflect.Descriptor instead.
func (*GetTopMinersRequest) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{5}
}

func (x *GetTopMinersRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type GetTopMinersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TopMiners []*MinerScore `protobuf:"bytes,1,rep,name=top_miners,json=topMiners,proto3" json:"top_miners,omitempty"`
}

func (x *GetTopMinersResponse) Reset() {
	*x = GetTopMinersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_reputation_rpc_rpc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTopMinersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTopMinersResponse) ProtoMessage() {}

func (x *GetTopMinersResponse) ProtoReflect() protoreflect.Message {
	mi := &file_reputation_rpc_rpc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTopMinersResponse.ProtoReflect.Descriptor instead.
func (*GetTopMinersResponse) Descriptor() ([]byte, []int) {
	return file_reputation_rpc_rpc_proto_rawDescGZIP(), []int{6}
}

func (x *GetTopMinersResponse) GetTopMiners() []*MinerScore {
	if x != nil {
		return x.TopMiners
	}
	return nil
}

var File_reputation_rpc_rpc_proto protoreflect.FileDescriptor

var file_reputation_rpc_rpc_proto_rawDesc = []byte{
	0x0a, 0x18, 0x72, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x72, 0x70, 0x63,
	0x2f, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e, 0x72, 0x65, 0x70, 0x75,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x72, 0x70, 0x63, 0x22, 0x36, 0x0a, 0x0a, 0x4d, 0x69,
	0x6e, 0x65, 0x72, 0x53, 0x63, 0x6f, 0x72, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x64, 0x64, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x64, 0x64, 0x72, 0x12, 0x14, 0x0a, 0x05,
	0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x73, 0x63, 0x6f,
	0x72, 0x65, 0x22, 0xb4, 0x01, 0x0a, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x70, 0x73, 0x65, 0x74, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x74, 0x69, 0x70, 0x73, 0x65, 0x74, 0x6b, 0x65, 0x79, 0x12, 0x39, 0x0a, 0x06, 0x6d, 0x69,
	0x6e, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x72, 0x65, 0x70,
	0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x49, 0x6e, 0x64, 0x65,
	0x78, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x6d,
	0x69, 0x6e, 0x65, 0x72, 0x73, 0x1a, 0x52, 0x0a, 0x0b, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x2d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x72, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x6c, 0x61, 0x73, 0x68, 0x65, 0x73, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x21, 0x0a, 0x07, 0x53, 0x6c, 0x61,
	0x73, 0x68, 0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x04, 0x52, 0x06, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x73, 0x22, 0x38, 0x0a, 0x10,
	0x41, 0x64, 0x64, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x14, 0x0a, 0x05, 0x6d, 0x61, 0x64, 0x64, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x6d, 0x61, 0x64, 0x64, 0x72, 0x22, 0x13, 0x0a, 0x11, 0x41, 0x64, 0x64, 0x53, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x2b, 0x0a, 0x13, 0x47,
	0x65, 0x74, 0x54, 0x6f, 0x70, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x22, 0x51, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x54,
	0x6f, 0x70, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x39, 0x0a, 0x0a, 0x74, 0x6f, 0x70, 0x5f, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x18, 0x01,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x72, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x53, 0x63, 0x6f, 0x72, 0x65,
	0x52, 0x09, 0x74, 0x6f, 0x70, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x32, 0xbd, 0x01, 0x0a, 0x0a,
	0x52, 0x50, 0x43, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x52, 0x0a, 0x09, 0x41, 0x64,
	0x64, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x20, 0x2e, 0x72, 0x65, 0x70, 0x75, 0x74, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x41, 0x64, 0x64, 0x53, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x72, 0x65, 0x70, 0x75,
	0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x41, 0x64, 0x64, 0x53, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x5b,
	0x0a, 0x0c, 0x47, 0x65, 0x74, 0x54, 0x6f, 0x70, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x12, 0x23,
	0x2e, 0x72, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x47, 0x65, 0x74, 0x54, 0x6f, 0x70, 0x4d, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x72, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x6f, 0x70, 0x4d, 0x69, 0x6e, 0x65, 0x72,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x10, 0x5a, 0x0e, 0x72,
	0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_reputation_rpc_rpc_proto_rawDescOnce sync.Once
	file_reputation_rpc_rpc_proto_rawDescData = file_reputation_rpc_rpc_proto_rawDesc
)

func file_reputation_rpc_rpc_proto_rawDescGZIP() []byte {
	file_reputation_rpc_rpc_proto_rawDescOnce.Do(func() {
		file_reputation_rpc_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_reputation_rpc_rpc_proto_rawDescData)
	})
	return file_reputation_rpc_rpc_proto_rawDescData
}

var file_reputation_rpc_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_reputation_rpc_rpc_proto_goTypes = []interface{}{
	(*MinerScore)(nil),           // 0: reputation.rpc.MinerScore
	(*Index)(nil),                // 1: reputation.rpc.Index
	(*Slashes)(nil),              // 2: reputation.rpc.Slashes
	(*AddSourceRequest)(nil),     // 3: reputation.rpc.AddSourceRequest
	(*AddSourceResponse)(nil),    // 4: reputation.rpc.AddSourceResponse
	(*GetTopMinersRequest)(nil),  // 5: reputation.rpc.GetTopMinersRequest
	(*GetTopMinersResponse)(nil), // 6: reputation.rpc.GetTopMinersResponse
	nil,                          // 7: reputation.rpc.Index.MinersEntry
}
var file_reputation_rpc_rpc_proto_depIdxs = []int32{
	7, // 0: reputation.rpc.Index.miners:type_name -> reputation.rpc.Index.MinersEntry
	0, // 1: reputation.rpc.GetTopMinersResponse.top_miners:type_name -> reputation.rpc.MinerScore
	2, // 2: reputation.rpc.Index.MinersEntry.value:type_name -> reputation.rpc.Slashes
	3, // 3: reputation.rpc.RPCService.AddSource:input_type -> reputation.rpc.AddSourceRequest
	5, // 4: reputation.rpc.RPCService.GetTopMiners:input_type -> reputation.rpc.GetTopMinersRequest
	4, // 5: reputation.rpc.RPCService.AddSource:output_type -> reputation.rpc.AddSourceResponse
	6, // 6: reputation.rpc.RPCService.GetTopMiners:output_type -> reputation.rpc.GetTopMinersResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_reputation_rpc_rpc_proto_init() }
func file_reputation_rpc_rpc_proto_init() {
	if File_reputation_rpc_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_reputation_rpc_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MinerScore); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_reputation_rpc_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Index); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_reputation_rpc_rpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Slashes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_reputation_rpc_rpc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddSourceRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_reputation_rpc_rpc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddSourceResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_reputation_rpc_rpc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTopMinersRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_reputation_rpc_rpc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTopMinersResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_reputation_rpc_rpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_reputation_rpc_rpc_proto_goTypes,
		DependencyIndexes: file_reputation_rpc_rpc_proto_depIdxs,
		MessageInfos:      file_reputation_rpc_rpc_proto_msgTypes,
	}.Build()
	File_reputation_rpc_rpc_proto = out.File
	file_reputation_rpc_rpc_proto_rawDesc = nil
	file_reputation_rpc_rpc_proto_goTypes = nil
	file_reputation_rpc_rpc_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// RPCServiceClient is the client API for RPCService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type RPCServiceClient interface {
	AddSource(ctx context.Context, in *AddSourceRequest, opts ...grpc.CallOption) (*AddSourceResponse, error)
	GetTopMiners(ctx context.Context, in *GetTopMinersRequest, opts ...grpc.CallOption) (*GetTopMinersResponse, error)
}

type rPCServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRPCServiceClient(cc grpc.ClientConnInterface) RPCServiceClient {
	return &rPCServiceClient{cc}
}

func (c *rPCServiceClient) AddSource(ctx context.Context, in *AddSourceRequest, opts ...grpc.CallOption) (*AddSourceResponse, error) {
	out := new(AddSourceResponse)
	err := c.cc.Invoke(ctx, "/reputation.rpc.RPCService/AddSource", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCServiceClient) GetTopMiners(ctx context.Context, in *GetTopMinersRequest, opts ...grpc.CallOption) (*GetTopMinersResponse, error) {
	out := new(GetTopMinersResponse)
	err := c.cc.Invoke(ctx, "/reputation.rpc.RPCService/GetTopMiners", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RPCServiceServer is the server API for RPCService service.
type RPCServiceServer interface {
	AddSource(context.Context, *AddSourceRequest) (*AddSourceResponse, error)
	GetTopMiners(context.Context, *GetTopMinersRequest) (*GetTopMinersResponse, error)
}

// UnimplementedRPCServiceServer can be embedded to have forward compatible implementations.
type UnimplementedRPCServiceServer struct {
}

func (*UnimplementedRPCServiceServer) AddSource(context.Context, *AddSourceRequest) (*AddSourceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddSource not implemented")
}
func (*UnimplementedRPCServiceServer) GetTopMiners(context.Context, *GetTopMinersRequest) (*GetTopMinersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTopMiners not implemented")
}

func RegisterRPCServiceServer(s *grpc.Server, srv RPCServiceServer) {
	s.RegisterService(&_RPCService_serviceDesc, srv)
}

func _RPCService_AddSource_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddSourceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServiceServer).AddSource(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reputation.rpc.RPCService/AddSource",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServiceServer).AddSource(ctx, req.(*AddSourceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCService_GetTopMiners_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTopMinersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServiceServer).GetTopMiners(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/reputation.rpc.RPCService/GetTopMiners",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServiceServer).GetTopMiners(ctx, req.(*GetTopMinersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _RPCService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "reputation.rpc.RPCService",
	HandlerType: (*RPCServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddSource",
			Handler:    _RPCService_AddSource_Handler,
		},
		{
			MethodName: "GetTopMiners",
			Handler:    _RPCService_GetTopMiners_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "reputation/rpc/rpc.proto",
}
