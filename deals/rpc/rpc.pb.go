// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: deals/rpc/rpc.proto

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

type DealConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Miner      string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	EpochPrice uint64 `protobuf:"varint,2,opt,name=epochPrice,proto3" json:"epochPrice,omitempty"`
}

func (x *DealConfig) Reset() {
	*x = DealConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DealConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DealConfig) ProtoMessage() {}

func (x *DealConfig) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DealConfig.ProtoReflect.Descriptor instead.
func (*DealConfig) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *DealConfig) GetMiner() string {
	if x != nil {
		return x.Miner
	}
	return ""
}

func (x *DealConfig) GetEpochPrice() uint64 {
	if x != nil {
		return x.EpochPrice
	}
	return 0
}

type DealInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProposalCid   string `protobuf:"bytes,1,opt,name=proposalCid,proto3" json:"proposalCid,omitempty"`
	StateID       uint64 `protobuf:"varint,2,opt,name=stateID,proto3" json:"stateID,omitempty"`
	StateName     string `protobuf:"bytes,3,opt,name=stateName,proto3" json:"stateName,omitempty"`
	Miner         string `protobuf:"bytes,4,opt,name=miner,proto3" json:"miner,omitempty"`
	PieceCID      []byte `protobuf:"bytes,5,opt,name=pieceCID,proto3" json:"pieceCID,omitempty"`
	Size          uint64 `protobuf:"varint,6,opt,name=size,proto3" json:"size,omitempty"`
	PricePerEpoch uint64 `protobuf:"varint,7,opt,name=pricePerEpoch,proto3" json:"pricePerEpoch,omitempty"`
	Duration      uint64 `protobuf:"varint,8,opt,name=duration,proto3" json:"duration,omitempty"`
}

func (x *DealInfo) Reset() {
	*x = DealInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DealInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DealInfo) ProtoMessage() {}

func (x *DealInfo) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DealInfo.ProtoReflect.Descriptor instead.
func (*DealInfo) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *DealInfo) GetProposalCid() string {
	if x != nil {
		return x.ProposalCid
	}
	return ""
}

func (x *DealInfo) GetStateID() uint64 {
	if x != nil {
		return x.StateID
	}
	return 0
}

func (x *DealInfo) GetStateName() string {
	if x != nil {
		return x.StateName
	}
	return ""
}

func (x *DealInfo) GetMiner() string {
	if x != nil {
		return x.Miner
	}
	return ""
}

func (x *DealInfo) GetPieceCID() []byte {
	if x != nil {
		return x.PieceCID
	}
	return nil
}

func (x *DealInfo) GetSize() uint64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *DealInfo) GetPricePerEpoch() uint64 {
	if x != nil {
		return x.PricePerEpoch
	}
	return 0
}

func (x *DealInfo) GetDuration() uint64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

type StoreParams struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address     string        `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	DealConfigs []*DealConfig `protobuf:"bytes,2,rep,name=dealConfigs,proto3" json:"dealConfigs,omitempty"`
	Duration    uint64        `protobuf:"varint,3,opt,name=duration,proto3" json:"duration,omitempty"`
}

func (x *StoreParams) Reset() {
	*x = StoreParams{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreParams) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreParams) ProtoMessage() {}

func (x *StoreParams) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreParams.ProtoReflect.Descriptor instead.
func (*StoreParams) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{2}
}

func (x *StoreParams) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *StoreParams) GetDealConfigs() []*DealConfig {
	if x != nil {
		return x.DealConfigs
	}
	return nil
}

func (x *StoreParams) GetDuration() uint64 {
	if x != nil {
		return x.Duration
	}
	return 0
}

type StoreRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*StoreRequest_StoreParams
	//	*StoreRequest_Chunk
	Payload isStoreRequest_Payload `protobuf_oneof:"payload"`
}

func (x *StoreRequest) Reset() {
	*x = StoreRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreRequest) ProtoMessage() {}

func (x *StoreRequest) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreRequest.ProtoReflect.Descriptor instead.
func (*StoreRequest) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{3}
}

func (m *StoreRequest) GetPayload() isStoreRequest_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *StoreRequest) GetStoreParams() *StoreParams {
	if x, ok := x.GetPayload().(*StoreRequest_StoreParams); ok {
		return x.StoreParams
	}
	return nil
}

func (x *StoreRequest) GetChunk() []byte {
	if x, ok := x.GetPayload().(*StoreRequest_Chunk); ok {
		return x.Chunk
	}
	return nil
}

type isStoreRequest_Payload interface {
	isStoreRequest_Payload()
}

type StoreRequest_StoreParams struct {
	StoreParams *StoreParams `protobuf:"bytes,1,opt,name=storeParams,proto3,oneof"`
}

type StoreRequest_Chunk struct {
	Chunk []byte `protobuf:"bytes,2,opt,name=chunk,proto3,oneof"`
}

func (*StoreRequest_StoreParams) isStoreRequest_Payload() {}

func (*StoreRequest_Chunk) isStoreRequest_Payload() {}

type StoreReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DataCid      string        `protobuf:"bytes,1,opt,name=dataCid,proto3" json:"dataCid,omitempty"`
	ProposalCids []string      `protobuf:"bytes,2,rep,name=proposalCids,proto3" json:"proposalCids,omitempty"`
	FailedDeals  []*DealConfig `protobuf:"bytes,3,rep,name=failedDeals,proto3" json:"failedDeals,omitempty"`
}

func (x *StoreReply) Reset() {
	*x = StoreReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreReply) ProtoMessage() {}

func (x *StoreReply) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreReply.ProtoReflect.Descriptor instead.
func (*StoreReply) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{4}
}

func (x *StoreReply) GetDataCid() string {
	if x != nil {
		return x.DataCid
	}
	return ""
}

func (x *StoreReply) GetProposalCids() []string {
	if x != nil {
		return x.ProposalCids
	}
	return nil
}

func (x *StoreReply) GetFailedDeals() []*DealConfig {
	if x != nil {
		return x.FailedDeals
	}
	return nil
}

type WatchRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Proposals []string `protobuf:"bytes,1,rep,name=proposals,proto3" json:"proposals,omitempty"`
}

func (x *WatchRequest) Reset() {
	*x = WatchRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WatchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WatchRequest) ProtoMessage() {}

func (x *WatchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WatchRequest.ProtoReflect.Descriptor instead.
func (*WatchRequest) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{5}
}

func (x *WatchRequest) GetProposals() []string {
	if x != nil {
		return x.Proposals
	}
	return nil
}

type WatchReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DealInfo *DealInfo `protobuf:"bytes,1,opt,name=dealInfo,proto3" json:"dealInfo,omitempty"`
}

func (x *WatchReply) Reset() {
	*x = WatchReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WatchReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WatchReply) ProtoMessage() {}

func (x *WatchReply) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WatchReply.ProtoReflect.Descriptor instead.
func (*WatchReply) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{6}
}

func (x *WatchReply) GetDealInfo() *DealInfo {
	if x != nil {
		return x.DealInfo
	}
	return nil
}

type RetrieveRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Cid     string `protobuf:"bytes,2,opt,name=cid,proto3" json:"cid,omitempty"`
}

func (x *RetrieveRequest) Reset() {
	*x = RetrieveRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveRequest) ProtoMessage() {}

func (x *RetrieveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveRequest.ProtoReflect.Descriptor instead.
func (*RetrieveRequest) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{7}
}

func (x *RetrieveRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *RetrieveRequest) GetCid() string {
	if x != nil {
		return x.Cid
	}
	return ""
}

type RetrieveReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Chunk []byte `protobuf:"bytes,1,opt,name=chunk,proto3" json:"chunk,omitempty"`
}

func (x *RetrieveReply) Reset() {
	*x = RetrieveReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_deals_rpc_rpc_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RetrieveReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RetrieveReply) ProtoMessage() {}

func (x *RetrieveReply) ProtoReflect() protoreflect.Message {
	mi := &file_deals_rpc_rpc_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RetrieveReply.ProtoReflect.Descriptor instead.
func (*RetrieveReply) Descriptor() ([]byte, []int) {
	return file_deals_rpc_rpc_proto_rawDescGZIP(), []int{8}
}

func (x *RetrieveReply) GetChunk() []byte {
	if x != nil {
		return x.Chunk
	}
	return nil
}

var File_deals_rpc_rpc_proto protoreflect.FileDescriptor

var file_deals_rpc_rpc_proto_rawDesc = []byte{
	0x0a, 0x13, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x72, 0x70, 0x63, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63,
	0x22, 0x42, 0x0a, 0x0a, 0x44, 0x65, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x14,
	0x0a, 0x05, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d,
	0x69, 0x6e, 0x65, 0x72, 0x12, 0x1e, 0x0a, 0x0a, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x50,
	0x72, 0x69, 0x63, 0x65, 0x22, 0xec, 0x01, 0x0a, 0x08, 0x44, 0x65, 0x61, 0x6c, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x20, 0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x43, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c,
	0x43, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x61, 0x74, 0x65, 0x49, 0x44, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x73, 0x74, 0x61, 0x74, 0x65, 0x49, 0x44, 0x12, 0x1c, 0x0a,
	0x09, 0x73, 0x74, 0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x73, 0x74, 0x61, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6d,
	0x69, 0x6e, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d, 0x69, 0x6e, 0x65,
	0x72, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x69, 0x65, 0x63, 0x65, 0x43, 0x49, 0x44, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x08, 0x70, 0x69, 0x65, 0x63, 0x65, 0x43, 0x49, 0x44, 0x12, 0x12, 0x0a,
	0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x04, 0x52, 0x04, 0x73, 0x69, 0x7a,
	0x65, 0x12, 0x24, 0x0a, 0x0d, 0x70, 0x72, 0x69, 0x63, 0x65, 0x50, 0x65, 0x72, 0x45, 0x70, 0x6f,
	0x63, 0x68, 0x18, 0x07, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x70, 0x72, 0x69, 0x63, 0x65, 0x50,
	0x65, 0x72, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x22, 0x7c, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x61, 0x72, 0x61,
	0x6d, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x37, 0x0a, 0x0b,
	0x64, 0x65, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x15, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65,
	0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0b, 0x64, 0x65, 0x61, 0x6c, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x67, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x22, 0x6d, 0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x3a, 0x0a, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72,
	0x70, 0x63, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x48, 0x00,
	0x52, 0x0b, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x73, 0x12, 0x16, 0x0a,
	0x05, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x05,
	0x63, 0x68, 0x75, 0x6e, 0x6b, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x22, 0x83, 0x01, 0x0a, 0x0a, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12,
	0x18, 0x0a, 0x07, 0x64, 0x61, 0x74, 0x61, 0x43, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x64, 0x61, 0x74, 0x61, 0x43, 0x69, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x70, 0x72, 0x6f,
	0x70, 0x6f, 0x73, 0x61, 0x6c, 0x43, 0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x0c, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x43, 0x69, 0x64, 0x73, 0x12, 0x37, 0x0a,
	0x0b, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x44, 0x65, 0x61, 0x6c, 0x73, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x15, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x44,
	0x65, 0x61, 0x6c, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x0b, 0x66, 0x61, 0x69, 0x6c, 0x65,
	0x64, 0x44, 0x65, 0x61, 0x6c, 0x73, 0x22, 0x2c, 0x0a, 0x0c, 0x57, 0x61, 0x74, 0x63, 0x68, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73,
	0x61, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x70, 0x6f,
	0x73, 0x61, 0x6c, 0x73, 0x22, 0x3d, 0x0a, 0x0a, 0x57, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x12, 0x2f, 0x0a, 0x08, 0x64, 0x65, 0x61, 0x6c, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63,
	0x2e, 0x44, 0x65, 0x61, 0x6c, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x64, 0x65, 0x61, 0x6c, 0x49,
	0x6e, 0x66, 0x6f, 0x22, 0x3d, 0x0a, 0x0f, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x10, 0x0a, 0x03, 0x63, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x63,
	0x69, 0x64, 0x22, 0x25, 0x0a, 0x0d, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x05, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x32, 0xc5, 0x01, 0x0a, 0x03, 0x52, 0x50,
	0x43, 0x12, 0x3b, 0x0a, 0x05, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x17, 0x2e, 0x64, 0x65, 0x61,
	0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x28, 0x01, 0x12, 0x3b,
	0x0a, 0x05, 0x57, 0x61, 0x74, 0x63, 0x68, 0x12, 0x17, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e,
	0x72, 0x70, 0x63, 0x2e, 0x57, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x15, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x57, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x30, 0x01, 0x12, 0x44, 0x0a, 0x08, 0x52,
	0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x12, 0x1a, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e,
	0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x64, 0x65, 0x61, 0x6c, 0x73, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x52, 0x65, 0x74, 0x72, 0x69, 0x65, 0x76, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x30,
	0x01, 0x42, 0x05, 0x5a, 0x03, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_deals_rpc_rpc_proto_rawDescOnce sync.Once
	file_deals_rpc_rpc_proto_rawDescData = file_deals_rpc_rpc_proto_rawDesc
)

func file_deals_rpc_rpc_proto_rawDescGZIP() []byte {
	file_deals_rpc_rpc_proto_rawDescOnce.Do(func() {
		file_deals_rpc_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_deals_rpc_rpc_proto_rawDescData)
	})
	return file_deals_rpc_rpc_proto_rawDescData
}

var file_deals_rpc_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_deals_rpc_rpc_proto_goTypes = []interface{}{
	(*DealConfig)(nil),      // 0: deals.rpc.DealConfig
	(*DealInfo)(nil),        // 1: deals.rpc.DealInfo
	(*StoreParams)(nil),     // 2: deals.rpc.StoreParams
	(*StoreRequest)(nil),    // 3: deals.rpc.StoreRequest
	(*StoreReply)(nil),      // 4: deals.rpc.StoreReply
	(*WatchRequest)(nil),    // 5: deals.rpc.WatchRequest
	(*WatchReply)(nil),      // 6: deals.rpc.WatchReply
	(*RetrieveRequest)(nil), // 7: deals.rpc.RetrieveRequest
	(*RetrieveReply)(nil),   // 8: deals.rpc.RetrieveReply
}
var file_deals_rpc_rpc_proto_depIdxs = []int32{
	0, // 0: deals.rpc.StoreParams.dealConfigs:type_name -> deals.rpc.DealConfig
	2, // 1: deals.rpc.StoreRequest.storeParams:type_name -> deals.rpc.StoreParams
	0, // 2: deals.rpc.StoreReply.failedDeals:type_name -> deals.rpc.DealConfig
	1, // 3: deals.rpc.WatchReply.dealInfo:type_name -> deals.rpc.DealInfo
	3, // 4: deals.rpc.RPC.Store:input_type -> deals.rpc.StoreRequest
	5, // 5: deals.rpc.RPC.Watch:input_type -> deals.rpc.WatchRequest
	7, // 6: deals.rpc.RPC.Retrieve:input_type -> deals.rpc.RetrieveRequest
	4, // 7: deals.rpc.RPC.Store:output_type -> deals.rpc.StoreReply
	6, // 8: deals.rpc.RPC.Watch:output_type -> deals.rpc.WatchReply
	8, // 9: deals.rpc.RPC.Retrieve:output_type -> deals.rpc.RetrieveReply
	7, // [7:10] is the sub-list for method output_type
	4, // [4:7] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_deals_rpc_rpc_proto_init() }
func file_deals_rpc_rpc_proto_init() {
	if File_deals_rpc_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_deals_rpc_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DealConfig); i {
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
		file_deals_rpc_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DealInfo); i {
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
		file_deals_rpc_rpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreParams); i {
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
		file_deals_rpc_rpc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreRequest); i {
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
		file_deals_rpc_rpc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreReply); i {
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
		file_deals_rpc_rpc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WatchRequest); i {
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
		file_deals_rpc_rpc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WatchReply); i {
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
		file_deals_rpc_rpc_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveRequest); i {
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
		file_deals_rpc_rpc_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RetrieveReply); i {
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
	file_deals_rpc_rpc_proto_msgTypes[3].OneofWrappers = []interface{}{
		(*StoreRequest_StoreParams)(nil),
		(*StoreRequest_Chunk)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_deals_rpc_rpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_deals_rpc_rpc_proto_goTypes,
		DependencyIndexes: file_deals_rpc_rpc_proto_depIdxs,
		MessageInfos:      file_deals_rpc_rpc_proto_msgTypes,
	}.Build()
	File_deals_rpc_rpc_proto = out.File
	file_deals_rpc_rpc_proto_rawDesc = nil
	file_deals_rpc_rpc_proto_goTypes = nil
	file_deals_rpc_rpc_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// RPCClient is the client API for RPC service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type RPCClient interface {
	Store(ctx context.Context, opts ...grpc.CallOption) (RPC_StoreClient, error)
	Watch(ctx context.Context, in *WatchRequest, opts ...grpc.CallOption) (RPC_WatchClient, error)
	Retrieve(ctx context.Context, in *RetrieveRequest, opts ...grpc.CallOption) (RPC_RetrieveClient, error)
}

type rPCClient struct {
	cc grpc.ClientConnInterface
}

func NewRPCClient(cc grpc.ClientConnInterface) RPCClient {
	return &rPCClient{cc}
}

func (c *rPCClient) Store(ctx context.Context, opts ...grpc.CallOption) (RPC_StoreClient, error) {
	stream, err := c.cc.NewStream(ctx, &_RPC_serviceDesc.Streams[0], "/deals.rpc.RPC/Store", opts...)
	if err != nil {
		return nil, err
	}
	x := &rPCStoreClient{stream}
	return x, nil
}

type RPC_StoreClient interface {
	Send(*StoreRequest) error
	CloseAndRecv() (*StoreReply, error)
	grpc.ClientStream
}

type rPCStoreClient struct {
	grpc.ClientStream
}

func (x *rPCStoreClient) Send(m *StoreRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *rPCStoreClient) CloseAndRecv() (*StoreReply, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(StoreReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *rPCClient) Watch(ctx context.Context, in *WatchRequest, opts ...grpc.CallOption) (RPC_WatchClient, error) {
	stream, err := c.cc.NewStream(ctx, &_RPC_serviceDesc.Streams[1], "/deals.rpc.RPC/Watch", opts...)
	if err != nil {
		return nil, err
	}
	x := &rPCWatchClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type RPC_WatchClient interface {
	Recv() (*WatchReply, error)
	grpc.ClientStream
}

type rPCWatchClient struct {
	grpc.ClientStream
}

func (x *rPCWatchClient) Recv() (*WatchReply, error) {
	m := new(WatchReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *rPCClient) Retrieve(ctx context.Context, in *RetrieveRequest, opts ...grpc.CallOption) (RPC_RetrieveClient, error) {
	stream, err := c.cc.NewStream(ctx, &_RPC_serviceDesc.Streams[2], "/deals.rpc.RPC/Retrieve", opts...)
	if err != nil {
		return nil, err
	}
	x := &rPCRetrieveClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type RPC_RetrieveClient interface {
	Recv() (*RetrieveReply, error)
	grpc.ClientStream
}

type rPCRetrieveClient struct {
	grpc.ClientStream
}

func (x *rPCRetrieveClient) Recv() (*RetrieveReply, error) {
	m := new(RetrieveReply)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RPCServer is the server API for RPC service.
type RPCServer interface {
	Store(RPC_StoreServer) error
	Watch(*WatchRequest, RPC_WatchServer) error
	Retrieve(*RetrieveRequest, RPC_RetrieveServer) error
}

// UnimplementedRPCServer can be embedded to have forward compatible implementations.
type UnimplementedRPCServer struct {
}

func (*UnimplementedRPCServer) Store(RPC_StoreServer) error {
	return status.Errorf(codes.Unimplemented, "method Store not implemented")
}
func (*UnimplementedRPCServer) Watch(*WatchRequest, RPC_WatchServer) error {
	return status.Errorf(codes.Unimplemented, "method Watch not implemented")
}
func (*UnimplementedRPCServer) Retrieve(*RetrieveRequest, RPC_RetrieveServer) error {
	return status.Errorf(codes.Unimplemented, "method Retrieve not implemented")
}

func RegisterRPCServer(s *grpc.Server, srv RPCServer) {
	s.RegisterService(&_RPC_serviceDesc, srv)
}

func _RPC_Store_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RPCServer).Store(&rPCStoreServer{stream})
}

type RPC_StoreServer interface {
	SendAndClose(*StoreReply) error
	Recv() (*StoreRequest, error)
	grpc.ServerStream
}

type rPCStoreServer struct {
	grpc.ServerStream
}

func (x *rPCStoreServer) SendAndClose(m *StoreReply) error {
	return x.ServerStream.SendMsg(m)
}

func (x *rPCStoreServer) Recv() (*StoreRequest, error) {
	m := new(StoreRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _RPC_Watch_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(WatchRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RPCServer).Watch(m, &rPCWatchServer{stream})
}

type RPC_WatchServer interface {
	Send(*WatchReply) error
	grpc.ServerStream
}

type rPCWatchServer struct {
	grpc.ServerStream
}

func (x *rPCWatchServer) Send(m *WatchReply) error {
	return x.ServerStream.SendMsg(m)
}

func _RPC_Retrieve_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(RetrieveRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RPCServer).Retrieve(m, &rPCRetrieveServer{stream})
}

type RPC_RetrieveServer interface {
	Send(*RetrieveReply) error
	grpc.ServerStream
}

type rPCRetrieveServer struct {
	grpc.ServerStream
}

func (x *rPCRetrieveServer) Send(m *RetrieveReply) error {
	return x.ServerStream.SendMsg(m)
}

var _RPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "deals.rpc.RPC",
	HandlerType: (*RPCServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Store",
			Handler:       _RPC_Store_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "Watch",
			Handler:       _RPC_Watch_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Retrieve",
			Handler:       _RPC_Retrieve_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "deals/rpc/rpc.proto",
}
