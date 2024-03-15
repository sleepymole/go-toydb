// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: raft/raftpb/raft.proto

package raftpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// RaftClient is the client API for Raft service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RaftClient interface {
	SendMessages(ctx context.Context, opts ...grpc.CallOption) (Raft_SendMessagesClient, error)
}

type raftClient struct {
	cc grpc.ClientConnInterface
}

func NewRaftClient(cc grpc.ClientConnInterface) RaftClient {
	return &raftClient{cc}
}

func (c *raftClient) SendMessages(ctx context.Context, opts ...grpc.CallOption) (Raft_SendMessagesClient, error) {
	stream, err := c.cc.NewStream(ctx, &Raft_ServiceDesc.Streams[0], "/Raft/SendMessages", opts...)
	if err != nil {
		return nil, err
	}
	x := &raftSendMessagesClient{stream}
	return x, nil
}

type Raft_SendMessagesClient interface {
	Send(*Message) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type raftSendMessagesClient struct {
	grpc.ClientStream
}

func (x *raftSendMessagesClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *raftSendMessagesClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// RaftServer is the server API for Raft service.
// All implementations must embed UnimplementedRaftServer
// for forward compatibility
type RaftServer interface {
	SendMessages(Raft_SendMessagesServer) error
	mustEmbedUnimplementedRaftServer()
}

// UnimplementedRaftServer must be embedded to have forward compatible implementations.
type UnimplementedRaftServer struct {
}

func (UnimplementedRaftServer) SendMessages(Raft_SendMessagesServer) error {
	return status.Errorf(codes.Unimplemented, "method SendMessages not implemented")
}
func (UnimplementedRaftServer) mustEmbedUnimplementedRaftServer() {}

// UnsafeRaftServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RaftServer will
// result in compilation errors.
type UnsafeRaftServer interface {
	mustEmbedUnimplementedRaftServer()
}

func RegisterRaftServer(s grpc.ServiceRegistrar, srv RaftServer) {
	s.RegisterService(&Raft_ServiceDesc, srv)
}

func _Raft_SendMessages_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(RaftServer).SendMessages(&raftSendMessagesServer{stream})
}

type Raft_SendMessagesServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type raftSendMessagesServer struct {
	grpc.ServerStream
}

func (x *raftSendMessagesServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *raftSendMessagesServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Raft_ServiceDesc is the grpc.ServiceDesc for Raft service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Raft_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Raft",
	HandlerType: (*RaftServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendMessages",
			Handler:       _Raft_SendMessages_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "raft/raftpb/raft.proto",
}
