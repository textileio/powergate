// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.12.1
// source: health/rpc/rpc.proto

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

type Status int32

const (
	Status_Ok       Status = 0
	Status_Degraded Status = 1
	Status_Error    Status = 2
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "Ok",
		1: "Degraded",
		2: "Error",
	}
	Status_value = map[string]int32{
		"Ok":       0,
		"Degraded": 1,
		"Error":    2,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_health_rpc_rpc_proto_enumTypes[0].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_health_rpc_rpc_proto_enumTypes[0]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_health_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

type CheckRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *CheckRequest) Reset() {
	*x = CheckRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_health_rpc_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckRequest) ProtoMessage() {}

func (x *CheckRequest) ProtoReflect() protoreflect.Message {
	mi := &file_health_rpc_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckRequest.ProtoReflect.Descriptor instead.
func (*CheckRequest) Descriptor() ([]byte, []int) {
	return file_health_rpc_rpc_proto_rawDescGZIP(), []int{0}
}

type CheckReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status   Status   `protobuf:"varint,1,opt,name=status,proto3,enum=health.rpc.Status" json:"status,omitempty"`
	Messages []string `protobuf:"bytes,2,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *CheckReply) Reset() {
	*x = CheckReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_health_rpc_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckReply) ProtoMessage() {}

func (x *CheckReply) ProtoReflect() protoreflect.Message {
	mi := &file_health_rpc_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckReply.ProtoReflect.Descriptor instead.
func (*CheckReply) Descriptor() ([]byte, []int) {
	return file_health_rpc_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *CheckReply) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_Ok
}

func (x *CheckReply) GetMessages() []string {
	if x != nil {
		return x.Messages
	}
	return nil
}

var File_health_rpc_rpc_proto protoreflect.FileDescriptor

var file_health_rpc_rpc_proto_rawDesc = []byte{
	0x0a, 0x14, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x72, 0x70, 0x63,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x72,
	0x70, 0x63, 0x22, 0x0e, 0x0a, 0x0c, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x22, 0x54, 0x0a, 0x0a, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x70, 0x6c, 0x79,
	0x12, 0x2a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x12, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x2a, 0x29, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x06, 0x0a, 0x02, 0x4f, 0x6b, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x44, 0x65,
	0x67, 0x72, 0x61, 0x64, 0x65, 0x64, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f,
	0x72, 0x10, 0x02, 0x32, 0x42, 0x0a, 0x03, 0x52, 0x50, 0x43, 0x12, 0x3b, 0x0a, 0x05, 0x43, 0x68,
	0x65, 0x63, 0x6b, 0x12, 0x18, 0x2e, 0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x72, 0x70, 0x63,
	0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x68, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x0c, 0x5a, 0x0a, 0x68, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x2f, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_health_rpc_rpc_proto_rawDescOnce sync.Once
	file_health_rpc_rpc_proto_rawDescData = file_health_rpc_rpc_proto_rawDesc
)

func file_health_rpc_rpc_proto_rawDescGZIP() []byte {
	file_health_rpc_rpc_proto_rawDescOnce.Do(func() {
		file_health_rpc_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_health_rpc_rpc_proto_rawDescData)
	})
	return file_health_rpc_rpc_proto_rawDescData
}

var file_health_rpc_rpc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_health_rpc_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_health_rpc_rpc_proto_goTypes = []interface{}{
	(Status)(0),          // 0: health.rpc.Status
	(*CheckRequest)(nil), // 1: health.rpc.CheckRequest
	(*CheckReply)(nil),   // 2: health.rpc.CheckReply
}
var file_health_rpc_rpc_proto_depIdxs = []int32{
	0, // 0: health.rpc.CheckReply.status:type_name -> health.rpc.Status
	1, // 1: health.rpc.RPC.Check:input_type -> health.rpc.CheckRequest
	2, // 2: health.rpc.RPC.Check:output_type -> health.rpc.CheckReply
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_health_rpc_rpc_proto_init() }
func file_health_rpc_rpc_proto_init() {
	if File_health_rpc_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_health_rpc_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckRequest); i {
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
		file_health_rpc_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckReply); i {
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
			RawDescriptor: file_health_rpc_rpc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_health_rpc_rpc_proto_goTypes,
		DependencyIndexes: file_health_rpc_rpc_proto_depIdxs,
		EnumInfos:         file_health_rpc_rpc_proto_enumTypes,
		MessageInfos:      file_health_rpc_rpc_proto_msgTypes,
	}.Build()
	File_health_rpc_rpc_proto = out.File
	file_health_rpc_rpc_proto_rawDesc = nil
	file_health_rpc_rpc_proto_goTypes = nil
	file_health_rpc_rpc_proto_depIdxs = nil
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
	Check(ctx context.Context, in *CheckRequest, opts ...grpc.CallOption) (*CheckReply, error)
}

type rPCClient struct {
	cc grpc.ClientConnInterface
}

func NewRPCClient(cc grpc.ClientConnInterface) RPCClient {
	return &rPCClient{cc}
}

func (c *rPCClient) Check(ctx context.Context, in *CheckRequest, opts ...grpc.CallOption) (*CheckReply, error) {
	out := new(CheckReply)
	err := c.cc.Invoke(ctx, "/health.rpc.RPC/Check", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RPCServer is the server API for RPC service.
type RPCServer interface {
	Check(context.Context, *CheckRequest) (*CheckReply, error)
}

// UnimplementedRPCServer can be embedded to have forward compatible implementations.
type UnimplementedRPCServer struct {
}

func (*UnimplementedRPCServer) Check(context.Context, *CheckRequest) (*CheckReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Check not implemented")
}

func RegisterRPCServer(s *grpc.Server, srv RPCServer) {
	s.RegisterService(&_RPC_serviceDesc, srv)
}

func _RPC_Check_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RPCServer).Check(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/health.rpc.RPC/Check",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RPCServer).Check(ctx, req.(*CheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _RPC_serviceDesc = grpc.ServiceDesc{
	ServiceName: "health.rpc.RPC",
	HandlerType: (*RPCServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Check",
			Handler:    _RPC_Check_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "health/rpc/rpc.proto",
}
