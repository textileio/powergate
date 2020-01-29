// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ask.proto

package filecoin_ask_pb

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

type Query struct {
	MaxPrice             uint64   `protobuf:"varint,1,opt,name=MaxPrice,proto3" json:"MaxPrice,omitempty"`
	PieceSize            uint64   `protobuf:"varint,2,opt,name=PieceSize,proto3" json:"PieceSize,omitempty"`
	Limit                int32    `protobuf:"varint,3,opt,name=Limit,proto3" json:"Limit,omitempty"`
	Offset               int32    `protobuf:"varint,4,opt,name=Offset,proto3" json:"Offset,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Query) Reset()         { *m = Query{} }
func (m *Query) String() string { return proto.CompactTextString(m) }
func (*Query) ProtoMessage()    {}
func (*Query) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{0}
}

func (m *Query) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Query.Unmarshal(m, b)
}
func (m *Query) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Query.Marshal(b, m, deterministic)
}
func (m *Query) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Query.Merge(m, src)
}
func (m *Query) XXX_Size() int {
	return xxx_messageInfo_Query.Size(m)
}
func (m *Query) XXX_DiscardUnknown() {
	xxx_messageInfo_Query.DiscardUnknown(m)
}

var xxx_messageInfo_Query proto.InternalMessageInfo

func (m *Query) GetMaxPrice() uint64 {
	if m != nil {
		return m.MaxPrice
	}
	return 0
}

func (m *Query) GetPieceSize() uint64 {
	if m != nil {
		return m.PieceSize
	}
	return 0
}

func (m *Query) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

func (m *Query) GetOffset() int32 {
	if m != nil {
		return m.Offset
	}
	return 0
}

type StorageAsk struct {
	Price                uint64   `protobuf:"varint,1,opt,name=price,proto3" json:"price,omitempty"`
	MinPieceSize         uint64   `protobuf:"varint,2,opt,name=minPieceSize,proto3" json:"minPieceSize,omitempty"`
	Miner                string   `protobuf:"bytes,3,opt,name=miner,proto3" json:"miner,omitempty"`
	Timestamp            uint64   `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Expiry               uint64   `protobuf:"varint,5,opt,name=expiry,proto3" json:"expiry,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StorageAsk) Reset()         { *m = StorageAsk{} }
func (m *StorageAsk) String() string { return proto.CompactTextString(m) }
func (*StorageAsk) ProtoMessage()    {}
func (*StorageAsk) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{1}
}

func (m *StorageAsk) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StorageAsk.Unmarshal(m, b)
}
func (m *StorageAsk) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StorageAsk.Marshal(b, m, deterministic)
}
func (m *StorageAsk) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StorageAsk.Merge(m, src)
}
func (m *StorageAsk) XXX_Size() int {
	return xxx_messageInfo_StorageAsk.Size(m)
}
func (m *StorageAsk) XXX_DiscardUnknown() {
	xxx_messageInfo_StorageAsk.DiscardUnknown(m)
}

var xxx_messageInfo_StorageAsk proto.InternalMessageInfo

func (m *StorageAsk) GetPrice() uint64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *StorageAsk) GetMinPieceSize() uint64 {
	if m != nil {
		return m.MinPieceSize
	}
	return 0
}

func (m *StorageAsk) GetMiner() string {
	if m != nil {
		return m.Miner
	}
	return ""
}

func (m *StorageAsk) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *StorageAsk) GetExpiry() uint64 {
	if m != nil {
		return m.Expiry
	}
	return 0
}

type Index struct {
	LastUpdated          uint64                 `protobuf:"varint,1,opt,name=lastUpdated,proto3" json:"lastUpdated,omitempty"`
	StorageMedianPrice   uint64                 `protobuf:"varint,2,opt,name=storageMedianPrice,proto3" json:"storageMedianPrice,omitempty"`
	Storage              map[string]*StorageAsk `protobuf:"bytes,3,rep,name=storage,proto3" json:"storage,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *Index) Reset()         { *m = Index{} }
func (m *Index) String() string { return proto.CompactTextString(m) }
func (*Index) ProtoMessage()    {}
func (*Index) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{2}
}

func (m *Index) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Index.Unmarshal(m, b)
}
func (m *Index) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Index.Marshal(b, m, deterministic)
}
func (m *Index) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Index.Merge(m, src)
}
func (m *Index) XXX_Size() int {
	return xxx_messageInfo_Index.Size(m)
}
func (m *Index) XXX_DiscardUnknown() {
	xxx_messageInfo_Index.DiscardUnknown(m)
}

var xxx_messageInfo_Index proto.InternalMessageInfo

func (m *Index) GetLastUpdated() uint64 {
	if m != nil {
		return m.LastUpdated
	}
	return 0
}

func (m *Index) GetStorageMedianPrice() uint64 {
	if m != nil {
		return m.StorageMedianPrice
	}
	return 0
}

func (m *Index) GetStorage() map[string]*StorageAsk {
	if m != nil {
		return m.Storage
	}
	return nil
}

type GetRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{3}
}

func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (m *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(m, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

type GetReply struct {
	Index                *Index   `protobuf:"bytes,1,opt,name=index,proto3" json:"index,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetReply) Reset()         { *m = GetReply{} }
func (m *GetReply) String() string { return proto.CompactTextString(m) }
func (*GetReply) ProtoMessage()    {}
func (*GetReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{4}
}

func (m *GetReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetReply.Unmarshal(m, b)
}
func (m *GetReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetReply.Marshal(b, m, deterministic)
}
func (m *GetReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetReply.Merge(m, src)
}
func (m *GetReply) XXX_Size() int {
	return xxx_messageInfo_GetReply.Size(m)
}
func (m *GetReply) XXX_DiscardUnknown() {
	xxx_messageInfo_GetReply.DiscardUnknown(m)
}

var xxx_messageInfo_GetReply proto.InternalMessageInfo

func (m *GetReply) GetIndex() *Index {
	if m != nil {
		return m.Index
	}
	return nil
}

type QueryRequest struct {
	Query                *Query   `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *QueryRequest) Reset()         { *m = QueryRequest{} }
func (m *QueryRequest) String() string { return proto.CompactTextString(m) }
func (*QueryRequest) ProtoMessage()    {}
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{5}
}

func (m *QueryRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryRequest.Unmarshal(m, b)
}
func (m *QueryRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryRequest.Marshal(b, m, deterministic)
}
func (m *QueryRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryRequest.Merge(m, src)
}
func (m *QueryRequest) XXX_Size() int {
	return xxx_messageInfo_QueryRequest.Size(m)
}
func (m *QueryRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryRequest proto.InternalMessageInfo

func (m *QueryRequest) GetQuery() *Query {
	if m != nil {
		return m.Query
	}
	return nil
}

type QueryReply struct {
	Asks                 []*StorageAsk `protobuf:"bytes,1,rep,name=asks,proto3" json:"asks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *QueryReply) Reset()         { *m = QueryReply{} }
func (m *QueryReply) String() string { return proto.CompactTextString(m) }
func (*QueryReply) ProtoMessage()    {}
func (*QueryReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_a9005bad68e0db4f, []int{6}
}

func (m *QueryReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueryReply.Unmarshal(m, b)
}
func (m *QueryReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueryReply.Marshal(b, m, deterministic)
}
func (m *QueryReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryReply.Merge(m, src)
}
func (m *QueryReply) XXX_Size() int {
	return xxx_messageInfo_QueryReply.Size(m)
}
func (m *QueryReply) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryReply.DiscardUnknown(m)
}

var xxx_messageInfo_QueryReply proto.InternalMessageInfo

func (m *QueryReply) GetAsks() []*StorageAsk {
	if m != nil {
		return m.Asks
	}
	return nil
}

func init() {
	proto.RegisterType((*Query)(nil), "filecoin.ask.pb.Query")
	proto.RegisterType((*StorageAsk)(nil), "filecoin.ask.pb.StorageAsk")
	proto.RegisterType((*Index)(nil), "filecoin.ask.pb.Index")
	proto.RegisterMapType((map[string]*StorageAsk)(nil), "filecoin.ask.pb.Index.StorageEntry")
	proto.RegisterType((*GetRequest)(nil), "filecoin.ask.pb.GetRequest")
	proto.RegisterType((*GetReply)(nil), "filecoin.ask.pb.GetReply")
	proto.RegisterType((*QueryRequest)(nil), "filecoin.ask.pb.QueryRequest")
	proto.RegisterType((*QueryReply)(nil), "filecoin.ask.pb.QueryReply")
}

func init() { proto.RegisterFile("ask.proto", fileDescriptor_a9005bad68e0db4f) }

var fileDescriptor_a9005bad68e0db4f = []byte{
	// 464 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x53, 0x5d, 0x6b, 0x13, 0x41,
	0x14, 0x75, 0xb2, 0xd9, 0xda, 0xdc, 0x0d, 0x28, 0x83, 0x94, 0x35, 0x55, 0x08, 0xe3, 0x4b, 0x1e,
	0x64, 0xa5, 0xf1, 0xa5, 0x88, 0x45, 0x52, 0x88, 0xa5, 0x60, 0x31, 0x4e, 0x2b, 0x3e, 0x4f, 0x93,
	0x1b, 0x19, 0xf6, 0xb3, 0x3b, 0x13, 0xc9, 0xfa, 0x1f, 0x7c, 0xf1, 0x27, 0xf8, 0x0f, 0xfd, 0x07,
	0x32, 0x1f, 0x35, 0x31, 0x1f, 0xbe, 0xed, 0x39, 0xf7, 0x9e, 0x7b, 0xce, 0xdc, 0x9d, 0x81, 0x8e,
	0x50, 0x69, 0x52, 0xd5, 0xa5, 0x2e, 0xe9, 0xa3, 0xb9, 0xcc, 0x70, 0x5a, 0xca, 0x22, 0xb1, 0xdc,
	0x2d, 0x2b, 0x21, 0xfc, 0xb4, 0xc0, 0xba, 0xa1, 0x3d, 0x38, 0xbc, 0x12, 0xcb, 0x49, 0x2d, 0xa7,
	0x18, 0x93, 0x3e, 0x19, 0xb4, 0xf9, 0x5f, 0x4c, 0x9f, 0x41, 0x67, 0x22, 0x71, 0x8a, 0xd7, 0xf2,
	0x3b, 0xc6, 0x2d, 0x5b, 0x5c, 0x11, 0xf4, 0x09, 0x84, 0x1f, 0x64, 0x2e, 0x75, 0x1c, 0xf4, 0xc9,
	0x20, 0xe4, 0x0e, 0xd0, 0x23, 0x38, 0xf8, 0x38, 0x9f, 0x2b, 0xd4, 0x71, 0xdb, 0xd2, 0x1e, 0xb1,
	0x9f, 0x04, 0xe0, 0x5a, 0x97, 0xb5, 0xf8, 0x8a, 0x23, 0x95, 0x1a, 0x71, 0xb5, 0xe6, 0xe9, 0x00,
	0x65, 0xd0, 0xcd, 0x65, 0xb1, 0xe9, 0xf9, 0x0f, 0x67, 0x94, 0xb9, 0x2c, 0xb0, 0xb6, 0xb6, 0x1d,
	0xee, 0x80, 0x89, 0xaa, 0x65, 0x8e, 0x4a, 0x8b, 0xbc, 0xb2, 0xce, 0x6d, 0xbe, 0x22, 0x4c, 0x28,
	0x5c, 0x56, 0xb2, 0x6e, 0xe2, 0xd0, 0x96, 0x3c, 0x62, 0xbf, 0x09, 0x84, 0x97, 0xc5, 0x0c, 0x97,
	0xb4, 0x0f, 0x51, 0x26, 0x94, 0xfe, 0x5c, 0xcd, 0x84, 0xc6, 0x99, 0x4f, 0xb5, 0x4e, 0xd1, 0x04,
	0xa8, 0x72, 0xf9, 0xaf, 0x70, 0x26, 0x45, 0xe1, 0x56, 0xe6, 0x12, 0xee, 0xa8, 0xd0, 0x33, 0x78,
	0xe8, 0xd9, 0x38, 0xe8, 0x07, 0x83, 0x68, 0xf8, 0x22, 0xd9, 0xf8, 0x09, 0x89, 0xb5, 0x4e, 0xfc,
	0x56, 0xc6, 0x85, 0xae, 0x1b, 0x7e, 0xaf, 0xe9, 0x7d, 0x81, 0xee, 0x7a, 0x81, 0x3e, 0x86, 0x20,
	0xc5, 0xc6, 0x06, 0xeb, 0x70, 0xf3, 0x49, 0x4f, 0x20, 0xfc, 0x26, 0xb2, 0x85, 0xcb, 0x10, 0x0d,
	0x8f, 0xb7, 0xc6, 0xaf, 0xd6, 0xcd, 0x5d, 0xe7, 0x9b, 0xd6, 0x29, 0x61, 0x5d, 0x80, 0x0b, 0xd4,
	0x1c, 0xef, 0x16, 0xa8, 0x34, 0x3b, 0x85, 0x43, 0x8b, 0xaa, 0xac, 0xa1, 0x2f, 0x21, 0x94, 0x26,
	0x91, 0x35, 0x89, 0x86, 0x47, 0xbb, 0xf3, 0x72, 0xd7, 0xc4, 0xde, 0x42, 0xd7, 0xde, 0x20, 0x3f,
	0xc9, 0xa8, 0xef, 0x0c, 0xde, 0xab, 0x76, 0xdd, 0xae, 0x89, 0x9d, 0x01, 0x78, 0xb5, 0x71, 0x7e,
	0x05, 0x6d, 0xa1, 0x52, 0x15, 0x13, 0xbb, 0xa8, 0xff, 0x9e, 0xc4, 0x36, 0x0e, 0x7f, 0x10, 0x08,
	0x46, 0x93, 0x4b, 0xfa, 0x0e, 0x82, 0x0b, 0xd4, 0x74, 0x5b, 0xb1, 0x3a, 0x62, 0xef, 0xe9, 0xee,
	0x62, 0x95, 0x35, 0xec, 0x01, 0x1d, 0xdf, 0xbf, 0x83, 0xe7, 0x7b, 0xf2, 0xfa, 0x21, 0xc7, 0xfb,
	0xca, 0x76, 0xcc, 0xf9, 0x09, 0xf4, 0x64, 0x99, 0x68, 0x5c, 0x6a, 0x99, 0xe1, 0x66, 0xeb, 0x79,
	0xf4, 0xde, 0x13, 0x23, 0x95, 0x4e, 0xc8, 0xaf, 0x56, 0x70, 0x73, 0x33, 0xbe, 0x3d, 0xb0, 0x2f,
	0xf3, 0xf5, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x0f, 0x7d, 0xe8, 0xeb, 0xa6, 0x03, 0x00, 0x00,
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
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetReply, error)
	Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryReply, error)
}

type aPIClient struct {
	cc *grpc.ClientConn
}

func NewAPIClient(cc *grpc.ClientConn) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetReply, error) {
	out := new(GetReply)
	err := c.cc.Invoke(ctx, "/filecoin.ask.pb.API/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aPIClient) Query(ctx context.Context, in *QueryRequest, opts ...grpc.CallOption) (*QueryReply, error) {
	out := new(QueryReply)
	err := c.cc.Invoke(ctx, "/filecoin.ask.pb.API/Query", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// APIServer is the server API for API service.
type APIServer interface {
	Get(context.Context, *GetRequest) (*GetReply, error)
	Query(context.Context, *QueryRequest) (*QueryReply, error)
}

// UnimplementedAPIServer can be embedded to have forward compatible implementations.
type UnimplementedAPIServer struct {
}

func (*UnimplementedAPIServer) Get(ctx context.Context, req *GetRequest) (*GetReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (*UnimplementedAPIServer) Query(ctx context.Context, req *QueryRequest) (*QueryReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Query not implemented")
}

func RegisterAPIServer(s *grpc.Server, srv APIServer) {
	s.RegisterService(&_API_serviceDesc, srv)
}

func _API_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/filecoin.ask.pb.API/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _API_Query_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).Query(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/filecoin.ask.pb.API/Query",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).Query(ctx, req.(*QueryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _API_serviceDesc = grpc.ServiceDesc{
	ServiceName: "filecoin.ask.pb.API",
	HandlerType: (*APIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _API_Get_Handler,
		},
		{
			MethodName: "Query",
			Handler:    _API_Query_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ask.proto",
}
