// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: controller/api/services/v1/session_recording_service.proto

package services

import (
	context "context"
	httpbody "google.golang.org/genproto/googleapis/api/httpbody"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	SessionRecordingService_GetSessionRecording_FullMethodName   = "/controller.api.services.v1.SessionRecordingService/GetSessionRecording"
	SessionRecordingService_ListSessionRecordings_FullMethodName = "/controller.api.services.v1.SessionRecordingService/ListSessionRecordings"
	SessionRecordingService_Download_FullMethodName              = "/controller.api.services.v1.SessionRecordingService/Download"
)

// SessionRecordingServiceClient is the client API for SessionRecordingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type SessionRecordingServiceClient interface {
	// GetSessionRecording returns a stored Session recording if present. The provided request
	// must include the Session recording ID for the Session recording being retrieved,
	// or the ID of the Session that was recorded. If that ID is missing, malformed or reference a
	// non existing resource, an error is returned.
	GetSessionRecording(ctx context.Context, in *GetSessionRecordingRequest, opts ...grpc.CallOption) (*GetSessionRecordingResponse, error)
	// ListSessionRecordings lists all session recordings.
	// Session recordings are ordered by start_time descending (most recently started first).
	ListSessionRecordings(ctx context.Context, in *ListSessionRecordingsRequest, opts ...grpc.CallOption) (*ListSessionRecordingsResponse, error)
	// Download returns the contents of the specified resource in the specified mime type.
	// Supports both Session ID and Session recording ID for looking up a Session recording.
	// Supports both Connection ID and Connection recording ID to look up a Connection recording.
	// A Channel recording ID is required to look up a Channel recording.
	// The only supported mime type is "application/x-asciicast".
	Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (SessionRecordingService_DownloadClient, error)
}

type sessionRecordingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewSessionRecordingServiceClient(cc grpc.ClientConnInterface) SessionRecordingServiceClient {
	return &sessionRecordingServiceClient{cc}
}

func (c *sessionRecordingServiceClient) GetSessionRecording(ctx context.Context, in *GetSessionRecordingRequest, opts ...grpc.CallOption) (*GetSessionRecordingResponse, error) {
	out := new(GetSessionRecordingResponse)
	err := c.cc.Invoke(ctx, SessionRecordingService_GetSessionRecording_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sessionRecordingServiceClient) ListSessionRecordings(ctx context.Context, in *ListSessionRecordingsRequest, opts ...grpc.CallOption) (*ListSessionRecordingsResponse, error) {
	out := new(ListSessionRecordingsResponse)
	err := c.cc.Invoke(ctx, SessionRecordingService_ListSessionRecordings_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *sessionRecordingServiceClient) Download(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (SessionRecordingService_DownloadClient, error) {
	stream, err := c.cc.NewStream(ctx, &SessionRecordingService_ServiceDesc.Streams[0], SessionRecordingService_Download_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &sessionRecordingServiceDownloadClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SessionRecordingService_DownloadClient interface {
	Recv() (*httpbody.HttpBody, error)
	grpc.ClientStream
}

type sessionRecordingServiceDownloadClient struct {
	grpc.ClientStream
}

func (x *sessionRecordingServiceDownloadClient) Recv() (*httpbody.HttpBody, error) {
	m := new(httpbody.HttpBody)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SessionRecordingServiceServer is the server API for SessionRecordingService service.
// All implementations must embed UnimplementedSessionRecordingServiceServer
// for forward compatibility
type SessionRecordingServiceServer interface {
	// GetSessionRecording returns a stored Session recording if present. The provided request
	// must include the Session recording ID for the Session recording being retrieved,
	// or the ID of the Session that was recorded. If that ID is missing, malformed or reference a
	// non existing resource, an error is returned.
	GetSessionRecording(context.Context, *GetSessionRecordingRequest) (*GetSessionRecordingResponse, error)
	// ListSessionRecordings lists all session recordings.
	// Session recordings are ordered by start_time descending (most recently started first).
	ListSessionRecordings(context.Context, *ListSessionRecordingsRequest) (*ListSessionRecordingsResponse, error)
	// Download returns the contents of the specified resource in the specified mime type.
	// Supports both Session ID and Session recording ID for looking up a Session recording.
	// Supports both Connection ID and Connection recording ID to look up a Connection recording.
	// A Channel recording ID is required to look up a Channel recording.
	// The only supported mime type is "application/x-asciicast".
	Download(*DownloadRequest, SessionRecordingService_DownloadServer) error
	mustEmbedUnimplementedSessionRecordingServiceServer()
}

// UnimplementedSessionRecordingServiceServer must be embedded to have forward compatible implementations.
type UnimplementedSessionRecordingServiceServer struct {
}

func (UnimplementedSessionRecordingServiceServer) GetSessionRecording(context.Context, *GetSessionRecordingRequest) (*GetSessionRecordingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSessionRecording not implemented")
}
func (UnimplementedSessionRecordingServiceServer) ListSessionRecordings(context.Context, *ListSessionRecordingsRequest) (*ListSessionRecordingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListSessionRecordings not implemented")
}
func (UnimplementedSessionRecordingServiceServer) Download(*DownloadRequest, SessionRecordingService_DownloadServer) error {
	return status.Errorf(codes.Unimplemented, "method Download not implemented")
}
func (UnimplementedSessionRecordingServiceServer) mustEmbedUnimplementedSessionRecordingServiceServer() {
}

// UnsafeSessionRecordingServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to SessionRecordingServiceServer will
// result in compilation errors.
type UnsafeSessionRecordingServiceServer interface {
	mustEmbedUnimplementedSessionRecordingServiceServer()
}

func RegisterSessionRecordingServiceServer(s grpc.ServiceRegistrar, srv SessionRecordingServiceServer) {
	s.RegisterService(&SessionRecordingService_ServiceDesc, srv)
}

func _SessionRecordingService_GetSessionRecording_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSessionRecordingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionRecordingServiceServer).GetSessionRecording(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SessionRecordingService_GetSessionRecording_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionRecordingServiceServer).GetSessionRecording(ctx, req.(*GetSessionRecordingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SessionRecordingService_ListSessionRecordings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListSessionRecordingsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(SessionRecordingServiceServer).ListSessionRecordings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: SessionRecordingService_ListSessionRecordings_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(SessionRecordingServiceServer).ListSessionRecordings(ctx, req.(*ListSessionRecordingsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _SessionRecordingService_Download_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DownloadRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SessionRecordingServiceServer).Download(m, &sessionRecordingServiceDownloadServer{stream})
}

type SessionRecordingService_DownloadServer interface {
	Send(*httpbody.HttpBody) error
	grpc.ServerStream
}

type sessionRecordingServiceDownloadServer struct {
	grpc.ServerStream
}

func (x *sessionRecordingServiceDownloadServer) Send(m *httpbody.HttpBody) error {
	return x.ServerStream.SendMsg(m)
}

// SessionRecordingService_ServiceDesc is the grpc.ServiceDesc for SessionRecordingService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var SessionRecordingService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "controller.api.services.v1.SessionRecordingService",
	HandlerType: (*SessionRecordingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSessionRecording",
			Handler:    _SessionRecordingService_GetSessionRecording_Handler,
		},
		{
			MethodName: "ListSessionRecordings",
			Handler:    _SessionRecordingService_ListSessionRecordings_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Download",
			Handler:       _SessionRecordingService_Download_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "controller/api/services/v1/session_recording_service.proto",
}
