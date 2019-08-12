// Code generated by protoc-gen-go. DO NOT EDIT.
// source: hook.proto

package gitalypb

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

type PreReceiveHookRequest struct {
	Repository           *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	KeyId                string      `protobuf:"bytes,2,opt,name=key_id,json=keyId,proto3" json:"key_id,omitempty"`
	Protocol             string      `protobuf:"bytes,3,opt,name=protocol,proto3" json:"protocol,omitempty"`
	Refs                 []string    `protobuf:"bytes,4,rep,name=refs,proto3" json:"refs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *PreReceiveHookRequest) Reset()         { *m = PreReceiveHookRequest{} }
func (m *PreReceiveHookRequest) String() string { return proto.CompactTextString(m) }
func (*PreReceiveHookRequest) ProtoMessage()    {}
func (*PreReceiveHookRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eef30da1c11ee1b, []int{0}
}

func (m *PreReceiveHookRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PreReceiveHookRequest.Unmarshal(m, b)
}
func (m *PreReceiveHookRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PreReceiveHookRequest.Marshal(b, m, deterministic)
}
func (m *PreReceiveHookRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PreReceiveHookRequest.Merge(m, src)
}
func (m *PreReceiveHookRequest) XXX_Size() int {
	return xxx_messageInfo_PreReceiveHookRequest.Size(m)
}
func (m *PreReceiveHookRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PreReceiveHookRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PreReceiveHookRequest proto.InternalMessageInfo

func (m *PreReceiveHookRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *PreReceiveHookRequest) GetKeyId() string {
	if m != nil {
		return m.KeyId
	}
	return ""
}

func (m *PreReceiveHookRequest) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

func (m *PreReceiveHookRequest) GetRefs() []string {
	if m != nil {
		return m.Refs
	}
	return nil
}

type PreReceiveHookResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PreReceiveHookResponse) Reset()         { *m = PreReceiveHookResponse{} }
func (m *PreReceiveHookResponse) String() string { return proto.CompactTextString(m) }
func (*PreReceiveHookResponse) ProtoMessage()    {}
func (*PreReceiveHookResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eef30da1c11ee1b, []int{1}
}

func (m *PreReceiveHookResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PreReceiveHookResponse.Unmarshal(m, b)
}
func (m *PreReceiveHookResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PreReceiveHookResponse.Marshal(b, m, deterministic)
}
func (m *PreReceiveHookResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PreReceiveHookResponse.Merge(m, src)
}
func (m *PreReceiveHookResponse) XXX_Size() int {
	return xxx_messageInfo_PreReceiveHookResponse.Size(m)
}
func (m *PreReceiveHookResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PreReceiveHookResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PreReceiveHookResponse proto.InternalMessageInfo

type PostReceiveHookRequest struct {
	Repository           *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	KeyId                string      `protobuf:"bytes,2,opt,name=key_id,json=keyId,proto3" json:"key_id,omitempty"`
	Refs                 []string    `protobuf:"bytes,3,rep,name=refs,proto3" json:"refs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *PostReceiveHookRequest) Reset()         { *m = PostReceiveHookRequest{} }
func (m *PostReceiveHookRequest) String() string { return proto.CompactTextString(m) }
func (*PostReceiveHookRequest) ProtoMessage()    {}
func (*PostReceiveHookRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eef30da1c11ee1b, []int{2}
}

func (m *PostReceiveHookRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostReceiveHookRequest.Unmarshal(m, b)
}
func (m *PostReceiveHookRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostReceiveHookRequest.Marshal(b, m, deterministic)
}
func (m *PostReceiveHookRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostReceiveHookRequest.Merge(m, src)
}
func (m *PostReceiveHookRequest) XXX_Size() int {
	return xxx_messageInfo_PostReceiveHookRequest.Size(m)
}
func (m *PostReceiveHookRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PostReceiveHookRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PostReceiveHookRequest proto.InternalMessageInfo

func (m *PostReceiveHookRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *PostReceiveHookRequest) GetKeyId() string {
	if m != nil {
		return m.KeyId
	}
	return ""
}

func (m *PostReceiveHookRequest) GetRefs() []string {
	if m != nil {
		return m.Refs
	}
	return nil
}

type PostReceiveHookResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PostReceiveHookResponse) Reset()         { *m = PostReceiveHookResponse{} }
func (m *PostReceiveHookResponse) String() string { return proto.CompactTextString(m) }
func (*PostReceiveHookResponse) ProtoMessage()    {}
func (*PostReceiveHookResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eef30da1c11ee1b, []int{3}
}

func (m *PostReceiveHookResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostReceiveHookResponse.Unmarshal(m, b)
}
func (m *PostReceiveHookResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostReceiveHookResponse.Marshal(b, m, deterministic)
}
func (m *PostReceiveHookResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostReceiveHookResponse.Merge(m, src)
}
func (m *PostReceiveHookResponse) XXX_Size() int {
	return xxx_messageInfo_PostReceiveHookResponse.Size(m)
}
func (m *PostReceiveHookResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PostReceiveHookResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PostReceiveHookResponse proto.InternalMessageInfo

type UpdateHookRequest struct {
	Repository           *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	KeyId                string      `protobuf:"bytes,2,opt,name=key_id,json=keyId,proto3" json:"key_id,omitempty"`
	Ref                  string      `protobuf:"bytes,3,opt,name=ref,proto3" json:"ref,omitempty"`
	OldValue             string      `protobuf:"bytes,4,opt,name=old_value,json=oldValue,proto3" json:"old_value,omitempty"`
	NewValue             string      `protobuf:"bytes,5,opt,name=new_value,json=newValue,proto3" json:"new_value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *UpdateHookRequest) Reset()         { *m = UpdateHookRequest{} }
func (m *UpdateHookRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateHookRequest) ProtoMessage()    {}
func (*UpdateHookRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eef30da1c11ee1b, []int{4}
}

func (m *UpdateHookRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateHookRequest.Unmarshal(m, b)
}
func (m *UpdateHookRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateHookRequest.Marshal(b, m, deterministic)
}
func (m *UpdateHookRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateHookRequest.Merge(m, src)
}
func (m *UpdateHookRequest) XXX_Size() int {
	return xxx_messageInfo_UpdateHookRequest.Size(m)
}
func (m *UpdateHookRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateHookRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateHookRequest proto.InternalMessageInfo

func (m *UpdateHookRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *UpdateHookRequest) GetKeyId() string {
	if m != nil {
		return m.KeyId
	}
	return ""
}

func (m *UpdateHookRequest) GetRef() string {
	if m != nil {
		return m.Ref
	}
	return ""
}

func (m *UpdateHookRequest) GetOldValue() string {
	if m != nil {
		return m.OldValue
	}
	return ""
}

func (m *UpdateHookRequest) GetNewValue() string {
	if m != nil {
		return m.NewValue
	}
	return ""
}

type UpdateHookResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UpdateHookResponse) Reset()         { *m = UpdateHookResponse{} }
func (m *UpdateHookResponse) String() string { return proto.CompactTextString(m) }
func (*UpdateHookResponse) ProtoMessage()    {}
func (*UpdateHookResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3eef30da1c11ee1b, []int{5}
}

func (m *UpdateHookResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateHookResponse.Unmarshal(m, b)
}
func (m *UpdateHookResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateHookResponse.Marshal(b, m, deterministic)
}
func (m *UpdateHookResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateHookResponse.Merge(m, src)
}
func (m *UpdateHookResponse) XXX_Size() int {
	return xxx_messageInfo_UpdateHookResponse.Size(m)
}
func (m *UpdateHookResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateHookResponse.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateHookResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*PreReceiveHookRequest)(nil), "gitaly.PreReceiveHookRequest")
	proto.RegisterType((*PreReceiveHookResponse)(nil), "gitaly.PreReceiveHookResponse")
	proto.RegisterType((*PostReceiveHookRequest)(nil), "gitaly.PostReceiveHookRequest")
	proto.RegisterType((*PostReceiveHookResponse)(nil), "gitaly.PostReceiveHookResponse")
	proto.RegisterType((*UpdateHookRequest)(nil), "gitaly.UpdateHookRequest")
	proto.RegisterType((*UpdateHookResponse)(nil), "gitaly.UpdateHookResponse")
}

func init() { proto.RegisterFile("hook.proto", fileDescriptor_3eef30da1c11ee1b) }

var fileDescriptor_3eef30da1c11ee1b = []byte{
	// 373 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x92, 0x41, 0x6f, 0xa2, 0x50,
	0x10, 0xc7, 0x83, 0x28, 0x91, 0x71, 0x0f, 0xbb, 0x2f, 0xab, 0x8b, 0x6c, 0x76, 0x6b, 0x38, 0x71,
	0x29, 0xb4, 0xf6, 0x1b, 0xf4, 0xd4, 0xf6, 0x64, 0x68, 0x6a, 0xd2, 0x5e, 0x0c, 0xc2, 0xa8, 0x04,
	0xea, 0xd0, 0x07, 0x6a, 0xf8, 0x0e, 0x3d, 0xf5, 0xd2, 0xef, 0xd0, 0x8f, 0xd8, 0x53, 0x03, 0x0f,
	0x95, 0x14, 0x3d, 0x7a, 0x9b, 0x99, 0xdf, 0x9b, 0x79, 0xff, 0xf7, 0x9f, 0x07, 0xb0, 0x20, 0x0a,
	0xad, 0x98, 0x53, 0x4a, 0x4c, 0x99, 0x07, 0xa9, 0x1b, 0x65, 0xfa, 0x8f, 0x64, 0xe1, 0x72, 0xf4,
	0x45, 0xd5, 0x78, 0x93, 0xa0, 0x3b, 0xe2, 0xe8, 0xa0, 0x87, 0xc1, 0x1a, 0x6f, 0x88, 0x42, 0x07,
	0x5f, 0x56, 0x98, 0xa4, 0x6c, 0x08, 0xc0, 0x31, 0xa6, 0x24, 0x48, 0x89, 0x67, 0x9a, 0x34, 0x90,
	0xcc, 0xce, 0x90, 0x59, 0x62, 0x88, 0xe5, 0xec, 0x88, 0x53, 0x39, 0xc5, 0xba, 0xa0, 0x84, 0x98,
	0x4d, 0x02, 0x5f, 0x6b, 0x0c, 0x24, 0x53, 0x75, 0x5a, 0x21, 0x66, 0xb7, 0x3e, 0xd3, 0xa1, 0x5d,
	0xdc, 0xe6, 0x51, 0xa4, 0xc9, 0x05, 0xd8, 0xe5, 0x8c, 0x41, 0x93, 0xe3, 0x2c, 0xd1, 0x9a, 0x03,
	0xd9, 0x54, 0x9d, 0x22, 0x36, 0x34, 0xe8, 0x7d, 0xd7, 0x94, 0xc4, 0xb4, 0x4c, 0xd0, 0xd8, 0x40,
	0x6f, 0x44, 0x49, 0x7a, 0x5a, 0xb9, 0x5b, 0x49, 0x72, 0x45, 0x52, 0x1f, 0xfe, 0xd4, 0x2e, 0x2e,
	0x35, 0x7d, 0x48, 0xf0, 0xeb, 0x21, 0xf6, 0xdd, 0xf4, 0x54, 0x7a, 0x7e, 0x82, 0xcc, 0x71, 0x56,
	0x3a, 0x97, 0x87, 0xec, 0x2f, 0xa8, 0x14, 0xf9, 0x93, 0xb5, 0x1b, 0xad, 0x50, 0x6b, 0x0a, 0x47,
	0x29, 0xf2, 0xc7, 0x79, 0x9e, 0xc3, 0x25, 0x6e, 0x4a, 0xd8, 0x12, 0x70, 0x89, 0x9b, 0x02, 0x1a,
	0xbf, 0x81, 0x55, 0xb5, 0x8a, 0x27, 0x0c, 0x5f, 0x1b, 0xd0, 0xc9, 0x0b, 0xf7, 0xc8, 0xd7, 0x81,
	0x87, 0x6c, 0x0c, 0xb0, 0x5f, 0x00, 0xfb, 0xb7, 0x95, 0x7d, 0xf0, 0xa3, 0xe8, 0xff, 0x8f, 0xe1,
	0xd2, 0x1f, 0xf5, 0xf3, 0xdd, 0x6c, 0xb5, 0x25, 0x5d, 0xba, 0x64, 0x8f, 0xd0, 0xa9, 0xb8, 0xc8,
	0xf6, 0x9d, 0x07, 0x77, 0xaa, 0x9f, 0x1d, 0xe5, 0xf5, 0xd1, 0x77, 0xa0, 0x88, 0x87, 0xb1, 0xfe,
	0xb6, 0xab, 0xb6, 0x14, 0x5d, 0x3f, 0x84, 0x6a, 0xb3, 0xae, 0x2f, 0x9e, 0xf2, 0x73, 0x91, 0x3b,
	0xb5, 0x3c, 0x7a, 0xb6, 0x45, 0x78, 0x4e, 0x7c, 0x6e, 0x8b, 0x6e, 0xbb, 0xf8, 0xbc, 0xf6, 0x9c,
	0xca, 0x3c, 0x9e, 0x4e, 0x95, 0xa2, 0x74, 0xf5, 0x15, 0x00, 0x00, 0xff, 0xff, 0xcf, 0x03, 0xc2,
	0x5c, 0x71, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// HookServiceClient is the client API for HookService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type HookServiceClient interface {
	PreReceive(ctx context.Context, in *PreReceiveHookRequest, opts ...grpc.CallOption) (*PreReceiveHookResponse, error)
	PostReceive(ctx context.Context, in *PostReceiveHookRequest, opts ...grpc.CallOption) (*PostReceiveHookResponse, error)
	Update(ctx context.Context, in *UpdateHookRequest, opts ...grpc.CallOption) (*UpdateHookResponse, error)
}

type hookServiceClient struct {
	cc *grpc.ClientConn
}

func NewHookServiceClient(cc *grpc.ClientConn) HookServiceClient {
	return &hookServiceClient{cc}
}

func (c *hookServiceClient) PreReceive(ctx context.Context, in *PreReceiveHookRequest, opts ...grpc.CallOption) (*PreReceiveHookResponse, error) {
	out := new(PreReceiveHookResponse)
	err := c.cc.Invoke(ctx, "/gitaly.HookService/PreReceive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hookServiceClient) PostReceive(ctx context.Context, in *PostReceiveHookRequest, opts ...grpc.CallOption) (*PostReceiveHookResponse, error) {
	out := new(PostReceiveHookResponse)
	err := c.cc.Invoke(ctx, "/gitaly.HookService/PostReceive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *hookServiceClient) Update(ctx context.Context, in *UpdateHookRequest, opts ...grpc.CallOption) (*UpdateHookResponse, error) {
	out := new(UpdateHookResponse)
	err := c.cc.Invoke(ctx, "/gitaly.HookService/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// HookServiceServer is the server API for HookService service.
type HookServiceServer interface {
	PreReceive(context.Context, *PreReceiveHookRequest) (*PreReceiveHookResponse, error)
	PostReceive(context.Context, *PostReceiveHookRequest) (*PostReceiveHookResponse, error)
	Update(context.Context, *UpdateHookRequest) (*UpdateHookResponse, error)
}

// UnimplementedHookServiceServer can be embedded to have forward compatible implementations.
type UnimplementedHookServiceServer struct {
}

func (*UnimplementedHookServiceServer) PreReceive(ctx context.Context, req *PreReceiveHookRequest) (*PreReceiveHookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PreReceive not implemented")
}
func (*UnimplementedHookServiceServer) PostReceive(ctx context.Context, req *PostReceiveHookRequest) (*PostReceiveHookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostReceive not implemented")
}
func (*UnimplementedHookServiceServer) Update(ctx context.Context, req *UpdateHookRequest) (*UpdateHookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func RegisterHookServiceServer(s *grpc.Server, srv HookServiceServer) {
	s.RegisterService(&_HookService_serviceDesc, srv)
}

func _HookService_PreReceive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PreReceiveHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HookServiceServer).PreReceive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.HookService/PreReceive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HookServiceServer).PreReceive(ctx, req.(*PreReceiveHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HookService_PostReceive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostReceiveHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HookServiceServer).PostReceive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.HookService/PostReceive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HookServiceServer).PostReceive(ctx, req.(*PostReceiveHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _HookService_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateHookRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HookServiceServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.HookService/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HookServiceServer).Update(ctx, req.(*UpdateHookRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _HookService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.HookService",
	HandlerType: (*HookServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PreReceive",
			Handler:    _HookService_PreReceive_Handler,
		},
		{
			MethodName: "PostReceive",
			Handler:    _HookService_PostReceive_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _HookService_Update_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "hook.proto",
}
