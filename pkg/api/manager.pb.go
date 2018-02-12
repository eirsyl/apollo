// Code generated by protoc-gen-go. DO NOT EDIT.
// source: manager.proto

package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type HealthRequest struct {
	InstanceId string `protobuf:"bytes,1,opt,name=instanceId" json:"instanceId,omitempty"`
}

func (m *HealthRequest) Reset()                    { *m = HealthRequest{} }
func (m *HealthRequest) String() string            { return proto.CompactTextString(m) }
func (*HealthRequest) ProtoMessage()               {}
func (*HealthRequest) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *HealthRequest) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

type HealthResponse struct {
	Ack bool `protobuf:"varint,1,opt,name=ack" json:"ack,omitempty"`
}

func (m *HealthResponse) Reset()                    { *m = HealthResponse{} }
func (m *HealthResponse) String() string            { return proto.CompactTextString(m) }
func (*HealthResponse) ProtoMessage()               {}
func (*HealthResponse) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{1} }

func (m *HealthResponse) GetAck() bool {
	if m != nil {
		return m.Ack
	}
	return false
}

func init() {
	proto.RegisterType((*HealthRequest)(nil), "api.HealthRequest")
	proto.RegisterType((*HealthResponse)(nil), "api.HealthResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Manager service

type ManagerClient interface {
	AgentHealth(ctx context.Context, opts ...grpc.CallOption) (Manager_AgentHealthClient, error)
}

type managerClient struct {
	cc *grpc.ClientConn
}

func NewManagerClient(cc *grpc.ClientConn) ManagerClient {
	return &managerClient{cc}
}

func (c *managerClient) AgentHealth(ctx context.Context, opts ...grpc.CallOption) (Manager_AgentHealthClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Manager_serviceDesc.Streams[0], c.cc, "/api.Manager/AgentHealth", opts...)
	if err != nil {
		return nil, err
	}
	x := &managerAgentHealthClient{stream}
	return x, nil
}

type Manager_AgentHealthClient interface {
	Send(*HealthRequest) error
	Recv() (*HealthResponse, error)
	grpc.ClientStream
}

type managerAgentHealthClient struct {
	grpc.ClientStream
}

func (x *managerAgentHealthClient) Send(m *HealthRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *managerAgentHealthClient) Recv() (*HealthResponse, error) {
	m := new(HealthResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Manager service

type ManagerServer interface {
	AgentHealth(Manager_AgentHealthServer) error
}

func RegisterManagerServer(s *grpc.Server, srv ManagerServer) {
	s.RegisterService(&_Manager_serviceDesc, srv)
}

func _Manager_AgentHealth_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ManagerServer).AgentHealth(&managerAgentHealthServer{stream})
}

type Manager_AgentHealthServer interface {
	Send(*HealthResponse) error
	Recv() (*HealthRequest, error)
	grpc.ServerStream
}

type managerAgentHealthServer struct {
	grpc.ServerStream
}

func (x *managerAgentHealthServer) Send(m *HealthResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *managerAgentHealthServer) Recv() (*HealthRequest, error) {
	m := new(HealthRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Manager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.Manager",
	HandlerType: (*ManagerServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "AgentHealth",
			Handler:       _Manager_AgentHealth_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "manager.proto",
}

func init() { proto.RegisterFile("manager.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 154 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcd, 0x4d, 0xcc, 0x4b,
	0x4c, 0x4f, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4e, 0x2c, 0xc8, 0x54, 0xd2,
	0xe7, 0xe2, 0xf5, 0x48, 0x4d, 0xcc, 0x29, 0xc9, 0x08, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x11,
	0x92, 0xe3, 0xe2, 0xca, 0xcc, 0x2b, 0x2e, 0x49, 0xcc, 0x4b, 0x4e, 0xf5, 0x4c, 0x91, 0x60, 0x54,
	0x60, 0xd4, 0xe0, 0x0c, 0x42, 0x12, 0x51, 0x52, 0xe2, 0xe2, 0x83, 0x69, 0x28, 0x2e, 0xc8, 0xcf,
	0x2b, 0x4e, 0x15, 0x12, 0xe0, 0x62, 0x4e, 0x4c, 0xce, 0x06, 0x2b, 0xe5, 0x08, 0x02, 0x31, 0x8d,
	0xdc, 0xb9, 0xd8, 0x7d, 0x21, 0x56, 0x09, 0xd9, 0x70, 0x71, 0x3b, 0xa6, 0xa7, 0xe6, 0x95, 0x40,
	0xf4, 0x08, 0x09, 0xe9, 0x25, 0x16, 0x64, 0xea, 0xa1, 0xd8, 0x28, 0x25, 0x8c, 0x22, 0x06, 0x31,
	0x54, 0x89, 0x41, 0x83, 0xd1, 0x80, 0x31, 0x89, 0x0d, 0xec, 0x52, 0x63, 0x40, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x36, 0x37, 0xf9, 0xd2, 0xba, 0x00, 0x00, 0x00,
}