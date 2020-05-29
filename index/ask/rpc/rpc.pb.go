// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: index/ask/rpc/rpc.proto

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

type Query struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MaxPrice  uint64 `protobuf:"varint,1,opt,name=max_price,json=maxPrice,proto3" json:"max_price,omitempty"`
	PieceSize uint64 `protobuf:"varint,2,opt,name=piece_size,json=pieceSize,proto3" json:"piece_size,omitempty"`
	Limit     int32  `protobuf:"varint,3,opt,name=limit,proto3" json:"limit,omitempty"`
	Offset    int32  `protobuf:"varint,4,opt,name=offset,proto3" json:"offset,omitempty"`
}

func (x *Query) Reset() {
	*x = Query{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Query) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Query) ProtoMessage() {}

func (x *Query) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Query.ProtoReflect.Descriptor instead.
func (*Query) Descriptor() ([]byte, []int) {
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *Query) GetMaxPrice() uint64 {
	if x != nil {
		return x.MaxPrice
	}
	return 0
}

func (x *Query) GetPieceSize() uint64 {
	if x != nil {
		return x.PieceSize
	}
	return 0
}

func (x *Query) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *Query) GetOffset() int32 {
	if x != nil {
		return x.Offset
	}
	return 0
}

type StorageAsk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Price        uint64 `protobuf:"varint,1,opt,name=price,proto3" json:"price,omitempty"`
	MinPieceSize uint64 `protobuf:"varint,2,opt,name=min_piece_size,json=minPieceSize,proto3" json:"min_piece_size,omitempty"`
	Miner        string `protobuf:"bytes,3,opt,name=miner,proto3" json:"miner,omitempty"`
	Timestamp    int64  `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Expiry       int64  `protobuf:"varint,5,opt,name=expiry,proto3" json:"expiry,omitempty"`
}

func (x *StorageAsk) Reset() {
	*x = StorageAsk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StorageAsk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StorageAsk) ProtoMessage() {}

func (x *StorageAsk) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StorageAsk.ProtoReflect.Descriptor instead.
func (*StorageAsk) Descriptor() ([]byte, []int) {
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *StorageAsk) GetPrice() uint64 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *StorageAsk) GetMinPieceSize() uint64 {
	if x != nil {
		return x.MinPieceSize
	}
	return 0
}

func (x *StorageAsk) GetMiner() string {
	if x != nil {
		return x.Miner
	}
	return ""
}

func (x *StorageAsk) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *StorageAsk) GetExpiry() int64 {
	if x != nil {
		return x.Expiry
	}
	return 0
}

type Index struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LastUpdated        int64                  `protobuf:"varint,1,opt,name=last_updated,json=lastUpdated,proto3" json:"last_updated,omitempty"`
	StorageMedianPrice uint64                 `protobuf:"varint,2,opt,name=storage_median_price,json=storageMedianPrice,proto3" json:"storage_median_price,omitempty"`
	Storage            map[string]*StorageAsk `protobuf:"bytes,3,rep,name=storage,proto3" json:"storage,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Index) Reset() {
	*x = Index{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Index) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Index) ProtoMessage() {}

func (x *Index) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[2]
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
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{2}
}

func (x *Index) GetLastUpdated() int64 {
	if x != nil {
		return x.LastUpdated
	}
	return 0
}

func (x *Index) GetStorageMedianPrice() uint64 {
	if x != nil {
		return x.StorageMedianPrice
	}
	return 0
}

func (x *Index) GetStorage() map[string]*StorageAsk {
	if x != nil {
		return x.Storage
	}
	return nil
}

type GetRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetRequest) Reset() {
	*x = GetRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRequest) ProtoMessage() {}

func (x *GetRequest) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRequest.ProtoReflect.Descriptor instead.
func (*GetRequest) Descriptor() ([]byte, []int) {
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{3}
}

type GetResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Index *Index `protobuf:"bytes,1,opt,name=index,proto3" json:"index,omitempty"`
}

func (x *GetResponse) Reset() {
	*x = GetResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetResponse) ProtoMessage() {}

func (x *GetResponse) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetResponse.ProtoReflect.Descriptor instead.
func (*GetResponse) Descriptor() ([]byte, []int) {
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{4}
}

func (x *GetResponse) GetIndex() *Index {
	if x != nil {
		return x.Index
	}
	return nil
}

type QueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Query *Query `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{5}
}

func (x *QueryRequest) GetQuery() *Query {
	if x != nil {
		return x.Query
	}
	return nil
}

type QueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Asks []*StorageAsk `protobuf:"bytes,1,rep,name=asks,proto3" json:"asks,omitempty"`
}

func (x *QueryResponse) Reset() {
	*x = QueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_index_ask_rpc_rpc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResponse) ProtoMessage() {}

func (x *QueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_index_ask_rpc_rpc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryResponse.ProtoReflect.Descriptor instead.
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return file_index_ask_rpc_rpc_proto_rawDescGZIP(), []int{6}
}

func (x *QueryResponse) GetAsks() []*StorageAsk {
	if x != nil {
		return x.Asks
	}
	return nil
}

var File_index_ask_rpc_rpc_proto protoreflect.FileDescriptor

var file_index_ask_rpc_rpc_proto_rawDesc = []byte{
	0x0a, 0x17, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2f, 0x61, 0x73, 0x6b, 0x2f, 0x72, 0x70, 0x63, 0x2f,
	0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x69, 0x6e, 0x64, 0x65, 0x78,
	0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63, 0x22, 0x71, 0x0a, 0x05, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x12, 0x1b, 0x0a, 0x09, 0x6d, 0x61, 0x78, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x52, 0x08, 0x6d, 0x61, 0x78, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x1d,
	0x0a, 0x0a, 0x70, 0x69, 0x65, 0x63, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x09, 0x70, 0x69, 0x65, 0x63, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x69,
	0x6d, 0x69, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x73, 0x65, 0x74, 0x22, 0x94, 0x01, 0x0a, 0x0a,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x41, 0x73, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x70, 0x72,
	0x69, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65,
	0x12, 0x24, 0x0a, 0x0e, 0x6d, 0x69, 0x6e, 0x5f, 0x70, 0x69, 0x65, 0x63, 0x65, 0x5f, 0x73, 0x69,
	0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0c, 0x6d, 0x69, 0x6e, 0x50, 0x69, 0x65,
	0x63, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x09, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78,
	0x70, 0x69, 0x72, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x65, 0x78, 0x70, 0x69,
	0x72, 0x79, 0x22, 0xf0, 0x01, 0x0a, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x21, 0x0a, 0x0c,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x12,
	0x30, 0x0a, 0x14, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x6d, 0x65, 0x64, 0x69, 0x61,
	0x6e, 0x5f, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x12, 0x73,
	0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x4d, 0x65, 0x64, 0x69, 0x61, 0x6e, 0x50, 0x72, 0x69, 0x63,
	0x65, 0x12, 0x3b, 0x0a, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x21, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72,
	0x70, 0x63, 0x2e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x1a, 0x55,
	0x0a, 0x0c, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x2f, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x41, 0x73, 0x6b, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x0c, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x39, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x2a, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70,
	0x63, 0x2e, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x22, 0x3a,
	0x0a, 0x0c, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2a,
	0x0a, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e,
	0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x51, 0x75,
	0x65, 0x72, 0x79, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79, 0x22, 0x3e, 0x0a, 0x0d, 0x51, 0x75,
	0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2d, 0x0a, 0x04, 0x61,
	0x73, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x41, 0x73, 0x6b, 0x52, 0x04, 0x61, 0x73, 0x6b, 0x73, 0x32, 0x92, 0x01, 0x0a, 0x0a, 0x52,
	0x50, 0x43, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3e, 0x0a, 0x03, 0x47, 0x65, 0x74,
	0x12, 0x19, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63,
	0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e, 0x69, 0x6e,
	0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x44, 0x0a, 0x05, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x12, 0x1b, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72,
	0x70, 0x63, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1c, 0x2e, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2e, 0x61, 0x73, 0x6b, 0x2e, 0x72, 0x70, 0x63, 0x2e,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42,
	0x0f, 0x5a, 0x0d, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x2f, 0x61, 0x73, 0x6b, 0x2f, 0x72, 0x70, 0x63,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_index_ask_rpc_rpc_proto_rawDescOnce sync.Once
	file_index_ask_rpc_rpc_proto_rawDescData = file_index_ask_rpc_rpc_proto_rawDesc
)

func file_index_ask_rpc_rpc_proto_rawDescGZIP() []byte {
	file_index_ask_rpc_rpc_proto_rawDescOnce.Do(func() {
		file_index_ask_rpc_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_index_ask_rpc_rpc_proto_rawDescData)
	})
	return file_index_ask_rpc_rpc_proto_rawDescData
}

var file_index_ask_rpc_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_index_ask_rpc_rpc_proto_goTypes = []interface{}{
	(*Query)(nil),         // 0: index.ask.rpc.Query
	(*StorageAsk)(nil),    // 1: index.ask.rpc.StorageAsk
	(*Index)(nil),         // 2: index.ask.rpc.Index
	(*GetRequest)(nil),    // 3: index.ask.rpc.GetRequest
	(*GetResponse)(nil),   // 4: index.ask.rpc.GetResponse
	(*QueryRequest)(nil),  // 5: index.ask.rpc.QueryRequest
	(*QueryResponse)(nil), // 6: index.ask.rpc.QueryResponse
	nil,                   // 7: index.ask.rpc.Index.StorageEntry
}
var file_index_ask_rpc_rpc_proto_depIdxs = []int32{
	7, // 0: index.ask.rpc.Index.storage:type_name -> index.ask.rpc.Index.StorageEntry
	2, // 1: index.ask.rpc.GetResponse.index:type_name -> index.ask.rpc.Index
	0, // 2: index.ask.rpc.QueryRequest.query:type_name -> index.ask.rpc.Query
	1, // 3: index.ask.rpc.QueryResponse.asks:type_name -> index.ask.rpc.StorageAsk
	1, // 4: index.ask.rpc.Index.StorageEntry.value:type_name -> index.ask.rpc.StorageAsk
	3, // 5: index.ask.rpc.RPCService.Get:input_type -> index.ask.rpc.GetRequest
	5, // 6: index.ask.rpc.RPCService.Query:input_type -> index.ask.rpc.QueryRequest
	4, // 7: index.ask.rpc.RPCService.Get:output_type -> index.ask.rpc.GetResponse
	6, // 8: index.ask.rpc.RPCService.Query:output_type -> index.ask.rpc.QueryResponse
	7, // [7:9] is the sub-list for method output_type
	5, // [5:7] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_index_ask_rpc_rpc_proto_init() }
func file_index_ask_rpc_rpc_proto_init() {
	if File_index_ask_rpc_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_index_ask_rpc_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Query); i {
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
		file_index_ask_rpc_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StorageAsk); i {
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
		file_index_ask_rpc_rpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_index_ask_rpc_rpc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRequest); i {
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
		file_index_ask_rpc_rpc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetResponse); i {
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
		file_index_ask_rpc_rpc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryRequest); i {
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
		file_index_ask_rpc_rpc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryResponse); i {
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
			RawDescriptor: file_index_ask_rpc_rpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_index_ask_rpc_rpc_proto_goTypes,
		DependencyIndexes: file_index_ask_rpc_rpc_proto_depIdxs,
		MessageInfos:      file_index_ask_rpc_rpc_proto_msgTypes,
	}.Build()
	File_index_ask_rpc_rpc_proto = out.File
	file_index_ask_rpc_rpc_proto_rawDesc = nil
	file_index_ask_rpc_rpc_proto_goTypes = nil
	file_index_ask_rpc_rpc_proto_depIdxs = nil
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
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error)
}

type rPCServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRPCServiceClient(cc grpc.ClientConnInterface) RPCServiceClient {
	return &rPCServiceClient{cc}
}

func (c *rPCServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/index.ask.rpc.RPCService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *rPCServiceClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryResponse, error) {
	out := new(QueryResponse)
	err := c.cc.Invoke(ctx, "/index.ask.rpc.RPCService/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RPCServiceServer is the server API for RPCService service.
type RPCServiceServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Query(context.Context, *QueryRequest) (*QueryResponse, error)
}

// UnimplementedRPCServiceServer can be embedded to have forward compatible implementations.
type UnimplementedRPCServiceServer struct {
}

func (*UnimplementedRPCServiceServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedRPCServiceServer) Query(context.Context, *QueryRequest) (*QueryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}

func RegisterRPCServiceServer(s *grpc.Server, srv RPCServiceServer) {
	s.RegisterService(&_RPCService_serviceDesc, srv)
}

func _RPCService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/index.ask.rpc.RPCService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RPCService_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServiceServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/index.ask.rpc.RPCService/Query",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServiceServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _RPCService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "index.ask.rpc.RPCService",
	HandlerType: (*RPCServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _RPCService_Get_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _RPCService_Query_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "index/ask/rpc/rpc.proto",
}
