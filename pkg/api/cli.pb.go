// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cli.proto

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	cli.proto
	manager.proto

It has these top-level messages:
	EmptyMessage
	StatusResponse
	NodesResponse
	DeleteNodeRequest
	DeleteNodeResponse
	StateRequest
	StateResponse
	NextExecutionRequest
	NextExecutionResponse
	ReportExecutionRequest
	ReportExecutionResponse
*/
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

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type EmptyMessage struct {
}

func (m *EmptyMessage) Reset()                    { *m = EmptyMessage{} }
func (m *EmptyMessage) String() string            { return proto.CompactTextString(m) }
func (*EmptyMessage) ProtoMessage()               {}
func (*EmptyMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type StatusResponse struct {
	Health int64                  `protobuf:"varint,1,opt,name=health" json:"health,omitempty"`
	State  int64                  `protobuf:"varint,2,opt,name=state" json:"state,omitempty"`
	Tasks  []*StatusResponse_Task `protobuf:"bytes,3,rep,name=tasks" json:"tasks,omitempty"`
}

func (m *StatusResponse) Reset()                    { *m = StatusResponse{} }
func (m *StatusResponse) String() string            { return proto.CompactTextString(m) }
func (*StatusResponse) ProtoMessage()               {}
func (*StatusResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *StatusResponse) GetHealth() int64 {
	if m != nil {
		return m.Health
	}
	return 0
}

func (m *StatusResponse) GetState() int64 {
	if m != nil {
		return m.State
	}
	return 0
}

func (m *StatusResponse) GetTasks() []*StatusResponse_Task {
	if m != nil {
		return m.Tasks
	}
	return nil
}

type StatusResponse_Task struct {
	Id       int64                          `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Status   int64                          `protobuf:"varint,2,opt,name=status" json:"status,omitempty"`
	Type     int64                          `protobuf:"varint,3,opt,name=type" json:"type,omitempty"`
	Commands []*StatusResponse_Task_Command `protobuf:"bytes,4,rep,name=commands" json:"commands,omitempty"`
}

func (m *StatusResponse_Task) Reset()                    { *m = StatusResponse_Task{} }
func (m *StatusResponse_Task) String() string            { return proto.CompactTextString(m) }
func (*StatusResponse_Task) ProtoMessage()               {}
func (*StatusResponse_Task) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

func (m *StatusResponse_Task) GetId() int64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *StatusResponse_Task) GetStatus() int64 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *StatusResponse_Task) GetType() int64 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *StatusResponse_Task) GetCommands() []*StatusResponse_Task_Command {
	if m != nil {
		return m.Commands
	}
	return nil
}

type StatusResponse_Task_Command struct {
	Id           string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Status       int64  `protobuf:"varint,2,opt,name=status" json:"status,omitempty"`
	Type         int64  `protobuf:"varint,3,opt,name=type" json:"type,omitempty"`
	NodeID       string `protobuf:"bytes,4,opt,name=nodeID" json:"nodeID,omitempty"`
	Retries      int64  `protobuf:"varint,5,opt,name=retries" json:"retries,omitempty"`
	Dependencies int64  `protobuf:"varint,6,opt,name=dependencies" json:"dependencies,omitempty"`
}

func (m *StatusResponse_Task_Command) Reset()         { *m = StatusResponse_Task_Command{} }
func (m *StatusResponse_Task_Command) String() string { return proto.CompactTextString(m) }
func (*StatusResponse_Task_Command) ProtoMessage()    {}
func (*StatusResponse_Task_Command) Descriptor() ([]byte, []int) {
	return fileDescriptor0, []int{1, 0, 0}
}

func (m *StatusResponse_Task_Command) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *StatusResponse_Task_Command) GetStatus() int64 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *StatusResponse_Task_Command) GetType() int64 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *StatusResponse_Task_Command) GetNodeID() string {
	if m != nil {
		return m.NodeID
	}
	return ""
}

func (m *StatusResponse_Task_Command) GetRetries() int64 {
	if m != nil {
		return m.Retries
	}
	return 0
}

func (m *StatusResponse_Task_Command) GetDependencies() int64 {
	if m != nil {
		return m.Dependencies
	}
	return 0
}

type NodesResponse struct {
	Nodes []*NodesResponse_Node `protobuf:"bytes,1,rep,name=nodes" json:"nodes,omitempty"`
}

func (m *NodesResponse) Reset()                    { *m = NodesResponse{} }
func (m *NodesResponse) String() string            { return proto.CompactTextString(m) }
func (*NodesResponse) ProtoMessage()               {}
func (*NodesResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *NodesResponse) GetNodes() []*NodesResponse_Node {
	if m != nil {
		return m.Nodes
	}
	return nil
}

type NodesResponse_Node struct {
	Id              string   `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Addr            string   `protobuf:"bytes,2,opt,name=addr" json:"addr,omitempty"`
	IsEmpty         bool     `protobuf:"varint,3,opt,name=isEmpty" json:"isEmpty,omitempty"`
	Nodes           []string `protobuf:"bytes,4,rep,name=nodes" json:"nodes,omitempty"`
	HostAnnotations []string `protobuf:"bytes,5,rep,name=hostAnnotations" json:"hostAnnotations,omitempty"`
	Online          bool     `protobuf:"varint,6,opt,name=online" json:"online,omitempty"`
}

func (m *NodesResponse_Node) Reset()                    { *m = NodesResponse_Node{} }
func (m *NodesResponse_Node) String() string            { return proto.CompactTextString(m) }
func (*NodesResponse_Node) ProtoMessage()               {}
func (*NodesResponse_Node) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2, 0} }

func (m *NodesResponse_Node) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *NodesResponse_Node) GetAddr() string {
	if m != nil {
		return m.Addr
	}
	return ""
}

func (m *NodesResponse_Node) GetIsEmpty() bool {
	if m != nil {
		return m.IsEmpty
	}
	return false
}

func (m *NodesResponse_Node) GetNodes() []string {
	if m != nil {
		return m.Nodes
	}
	return nil
}

func (m *NodesResponse_Node) GetHostAnnotations() []string {
	if m != nil {
		return m.HostAnnotations
	}
	return nil
}

func (m *NodesResponse_Node) GetOnline() bool {
	if m != nil {
		return m.Online
	}
	return false
}

type DeleteNodeRequest struct {
	NodeID string `protobuf:"bytes,1,opt,name=nodeID" json:"nodeID,omitempty"`
}

func (m *DeleteNodeRequest) Reset()                    { *m = DeleteNodeRequest{} }
func (m *DeleteNodeRequest) String() string            { return proto.CompactTextString(m) }
func (*DeleteNodeRequest) ProtoMessage()               {}
func (*DeleteNodeRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *DeleteNodeRequest) GetNodeID() string {
	if m != nil {
		return m.NodeID
	}
	return ""
}

type DeleteNodeResponse struct {
	Success bool `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
}

func (m *DeleteNodeResponse) Reset()                    { *m = DeleteNodeResponse{} }
func (m *DeleteNodeResponse) String() string            { return proto.CompactTextString(m) }
func (*DeleteNodeResponse) ProtoMessage()               {}
func (*DeleteNodeResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *DeleteNodeResponse) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func init() {
	proto.RegisterType((*EmptyMessage)(nil), "api.EmptyMessage")
	proto.RegisterType((*StatusResponse)(nil), "api.StatusResponse")
	proto.RegisterType((*StatusResponse_Task)(nil), "api.StatusResponse.Task")
	proto.RegisterType((*StatusResponse_Task_Command)(nil), "api.StatusResponse.Task.Command")
	proto.RegisterType((*NodesResponse)(nil), "api.NodesResponse")
	proto.RegisterType((*NodesResponse_Node)(nil), "api.NodesResponse.Node")
	proto.RegisterType((*DeleteNodeRequest)(nil), "api.DeleteNodeRequest")
	proto.RegisterType((*DeleteNodeResponse)(nil), "api.DeleteNodeResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for CLI service

type CLIClient interface {
	Status(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*StatusResponse, error)
	Nodes(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*NodesResponse, error)
	DeleteNode(ctx context.Context, in *DeleteNodeRequest, opts ...grpc.CallOption) (*DeleteNodeResponse, error)
}

type cLIClient struct {
	cc *grpc.ClientConn
}

func NewCLIClient(cc *grpc.ClientConn) CLIClient {
	return &cLIClient{cc}
}

func (c *cLIClient) Status(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := grpc.Invoke(ctx, "/api.CLI/Status", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cLIClient) Nodes(ctx context.Context, in *EmptyMessage, opts ...grpc.CallOption) (*NodesResponse, error) {
	out := new(NodesResponse)
	err := grpc.Invoke(ctx, "/api.CLI/Nodes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cLIClient) DeleteNode(ctx context.Context, in *DeleteNodeRequest, opts ...grpc.CallOption) (*DeleteNodeResponse, error) {
	out := new(DeleteNodeResponse)
	err := grpc.Invoke(ctx, "/api.CLI/DeleteNode", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for CLI service

type CLIServer interface {
	Status(context.Context, *EmptyMessage) (*StatusResponse, error)
	Nodes(context.Context, *EmptyMessage) (*NodesResponse, error)
	DeleteNode(context.Context, *DeleteNodeRequest) (*DeleteNodeResponse, error)
}

func RegisterCLIServer(s *grpc.Server, srv CLIServer) {
	s.RegisterService(&_CLI_serviceDesc, srv)
}

func _CLI_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CLIServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.CLI/Status",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CLIServer).Status(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _CLI_Nodes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CLIServer).Nodes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.CLI/Nodes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CLIServer).Nodes(ctx, req.(*EmptyMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _CLI_DeleteNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteNodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CLIServer).DeleteNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.CLI/DeleteNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CLIServer).DeleteNode(ctx, req.(*DeleteNodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _CLI_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.CLI",
	HandlerType: (*CLIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _CLI_Status_Handler,
		},
		{
			MethodName: "Nodes",
			Handler:    _CLI_Nodes_Handler,
		},
		{
			MethodName: "DeleteNode",
			Handler:    _CLI_DeleteNode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cli.proto",
}

func init() { proto.RegisterFile("cli.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 455 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0xae, 0xe3, 0x9f, 0x26, 0x43, 0x09, 0xea, 0x80, 0xd2, 0x95, 0x4f, 0x91, 0x4f, 0x91, 0x10,
	0x16, 0x0a, 0x57, 0x24, 0x84, 0x5a, 0x0e, 0x95, 0x80, 0x83, 0xe1, 0x05, 0x16, 0x7b, 0x44, 0xac,
	0x26, 0xbb, 0x26, 0xb3, 0x39, 0xf4, 0x29, 0x78, 0x00, 0x10, 0x4f, 0xc0, 0x2b, 0xf1, 0x2e, 0x68,
	0xc7, 0x1b, 0xea, 0x90, 0x72, 0xe0, 0xe6, 0x6f, 0xe6, 0x9b, 0x9f, 0xef, 0xdb, 0x31, 0x4c, 0xea,
	0x75, 0x5b, 0x76, 0x5b, 0xeb, 0x2c, 0xc6, 0xba, 0x6b, 0x8b, 0x29, 0x9c, 0xbd, 0xd9, 0x74, 0xee,
	0xf6, 0x1d, 0x31, 0xeb, 0xcf, 0x54, 0x7c, 0x8d, 0x61, 0xfa, 0xc1, 0x69, 0xb7, 0xe3, 0x8a, 0xb8,
	0xb3, 0x86, 0x09, 0x67, 0x90, 0xad, 0x48, 0xaf, 0xdd, 0x4a, 0x45, 0xf3, 0x68, 0x11, 0x57, 0x01,
	0xe1, 0x13, 0x48, 0xd9, 0x69, 0x47, 0x6a, 0x24, 0xe1, 0x1e, 0x60, 0x09, 0xa9, 0xd3, 0x7c, 0xc3,
	0x2a, 0x9e, 0xc7, 0x8b, 0x07, 0x4b, 0x55, 0xea, 0xae, 0x2d, 0x0f, 0x3b, 0x96, 0x1f, 0x35, 0xdf,
	0x54, 0x3d, 0x2d, 0xff, 0x31, 0x82, 0xc4, 0x63, 0x9c, 0xc2, 0xa8, 0x6d, 0xc2, 0x88, 0x51, 0xdb,
	0xf8, 0xb1, 0x2c, 0x65, 0xa1, 0x7f, 0x40, 0x88, 0x90, 0xb8, 0xdb, 0x8e, 0x54, 0x2c, 0x51, 0xf9,
	0xc6, 0x97, 0x30, 0xae, 0xed, 0x66, 0xa3, 0x4d, 0xc3, 0x2a, 0x91, 0xb9, 0xf3, 0x7f, 0xcd, 0x2d,
	0x2f, 0x7b, 0x62, 0xf5, 0xa7, 0x22, 0xff, 0x16, 0xc1, 0x69, 0x88, 0x0e, 0xb6, 0x98, 0xfc, 0xf7,
	0x16, 0x33, 0xc8, 0x8c, 0x6d, 0xe8, 0xfa, 0x4a, 0x25, 0x52, 0x1f, 0x10, 0x2a, 0x38, 0xdd, 0x92,
	0xdb, 0xb6, 0xc4, 0x2a, 0x15, 0xfa, 0x1e, 0x62, 0x01, 0x67, 0x0d, 0x75, 0x64, 0x1a, 0x32, 0xb5,
	0x4f, 0x67, 0x92, 0x3e, 0x88, 0x15, 0xbf, 0x22, 0x78, 0xf8, 0xde, 0x36, 0x74, 0xf7, 0x20, 0xcf,
	0x20, 0xf5, 0x9d, 0x59, 0x45, 0x22, 0xf5, 0x42, 0xa4, 0x1e, 0x50, 0x04, 0x55, 0x3d, 0x2b, 0xff,
	0x1e, 0x41, 0xe2, 0xf1, 0x91, 0x36, 0x84, 0x44, 0x37, 0xcd, 0x56, 0x94, 0x4d, 0x2a, 0xf9, 0xf6,
	0xbb, 0xb6, 0x2c, 0x17, 0x21, 0xd2, 0xc6, 0xd5, 0x1e, 0xfa, 0xe7, 0xee, 0xa7, 0x7a, 0x83, 0x27,
	0xa1, 0x39, 0x2e, 0xe0, 0xd1, 0xca, 0xb2, 0x7b, 0x6d, 0x8c, 0x75, 0xda, 0xb5, 0xd6, 0x78, 0x8d,
	0x3e, 0xff, 0x77, 0xd8, 0xbb, 0x63, 0xcd, 0xba, 0x35, 0x24, 0x2a, 0xc7, 0x55, 0x40, 0xc5, 0x53,
	0x38, 0xbf, 0xa2, 0x35, 0x39, 0x92, 0x9d, 0xe9, 0xcb, 0x8e, 0xd8, 0x0d, 0xac, 0x8c, 0x86, 0x56,
	0x16, 0x25, 0xe0, 0x90, 0x1c, 0x0c, 0x51, 0x70, 0xca, 0xbb, 0xba, 0x26, 0x66, 0xa1, 0x8f, 0xab,
	0x3d, 0x5c, 0xfe, 0x8c, 0x20, 0xbe, 0x7c, 0x7b, 0x8d, 0x4b, 0xc8, 0xfa, 0x5b, 0xc0, 0x73, 0x71,
	0x6b, 0x78, 0xf3, 0xf9, 0xe3, 0x7b, 0x6e, 0xa5, 0x38, 0xc1, 0xe7, 0x90, 0x8a, 0xa9, 0xf7, 0x95,
	0xe0, 0xb1, 0xe7, 0xc5, 0x09, 0xbe, 0x02, 0xb8, 0xdb, 0x0e, 0x67, 0xc2, 0x39, 0xd2, 0x96, 0x5f,
	0x1c, 0xc5, 0xf7, 0x0d, 0x3e, 0x65, 0xf2, 0x67, 0xbe, 0xf8, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x73,
	0x09, 0xb1, 0x00, 0xa6, 0x03, 0x00, 0x00,
}
