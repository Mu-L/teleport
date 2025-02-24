// Teleport
// Copyright (C) 2025 Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: teleport/lib/vnet/v1/client_application_service.proto

package vnetv1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ClientApplicationService_AuthenticateProcess_FullMethodName      = "/teleport.lib.vnet.v1.ClientApplicationService/AuthenticateProcess"
	ClientApplicationService_Ping_FullMethodName                     = "/teleport.lib.vnet.v1.ClientApplicationService/Ping"
	ClientApplicationService_ResolveAppInfo_FullMethodName           = "/teleport.lib.vnet.v1.ClientApplicationService/ResolveAppInfo"
	ClientApplicationService_ReissueAppCert_FullMethodName           = "/teleport.lib.vnet.v1.ClientApplicationService/ReissueAppCert"
	ClientApplicationService_SignForApp_FullMethodName               = "/teleport.lib.vnet.v1.ClientApplicationService/SignForApp"
	ClientApplicationService_OnNewConnection_FullMethodName          = "/teleport.lib.vnet.v1.ClientApplicationService/OnNewConnection"
	ClientApplicationService_OnInvalidLocalPort_FullMethodName       = "/teleport.lib.vnet.v1.ClientApplicationService/OnInvalidLocalPort"
	ClientApplicationService_GetTargetOSConfiguration_FullMethodName = "/teleport.lib.vnet.v1.ClientApplicationService/GetTargetOSConfiguration"
)

// ClientApplicationServiceClient is the client API for ClientApplicationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ClientApplicationService is a service the VNet client applications provide to
// the VNet admin process to facilate app queries, certificate issuance,
// metrics, error reporting, and signatures.
type ClientApplicationServiceClient interface {
	// AuthenticateProcess mutually authenticates client applicates to the admin
	// service.
	AuthenticateProcess(ctx context.Context, in *AuthenticateProcessRequest, opts ...grpc.CallOption) (*AuthenticateProcessResponse, error)
	// Ping is used by the admin process to regularly poll that the client
	// application is still running.
	Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	// ResolveAppInfo returns info for the given app fqdn, or an error if the app
	// is not present in any logged-in cluster.
	ResolveAppInfo(ctx context.Context, in *ResolveAppInfoRequest, opts ...grpc.CallOption) (*ResolveAppInfoResponse, error)
	// ReissueAppCert issues a new app cert.
	ReissueAppCert(ctx context.Context, in *ReissueAppCertRequest, opts ...grpc.CallOption) (*ReissueAppCertResponse, error)
	// SignForApp issues a signature with the private key associated with an x509
	// certificate previously issued for a requested app.
	SignForApp(ctx context.Context, in *SignForAppRequest, opts ...grpc.CallOption) (*SignForAppResponse, error)
	// OnNewConnection gets called whenever a new connection is about to be
	// established through VNet for observability.
	OnNewConnection(ctx context.Context, in *OnNewConnectionRequest, opts ...grpc.CallOption) (*OnNewConnectionResponse, error)
	// OnInvalidLocalPort gets called before VNet refuses to handle a connection
	// to a multi-port TCP app because the provided port does not match any of the
	// TCP ports in the app spec.
	OnInvalidLocalPort(ctx context.Context, in *OnInvalidLocalPortRequest, opts ...grpc.CallOption) (*OnInvalidLocalPortResponse, error)
	// GetTargetOSConfiguration gets the target OS configuration.
	GetTargetOSConfiguration(ctx context.Context, in *GetTargetOSConfigurationRequest, opts ...grpc.CallOption) (*GetTargetOSConfigurationResponse, error)
}

type clientApplicationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientApplicationServiceClient(cc grpc.ClientConnInterface) ClientApplicationServiceClient {
	return &clientApplicationServiceClient{cc}
}

func (c *clientApplicationServiceClient) AuthenticateProcess(ctx context.Context, in *AuthenticateProcessRequest, opts ...grpc.CallOption) (*AuthenticateProcessResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthenticateProcessResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_AuthenticateProcess_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(PingResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_Ping_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) ResolveAppInfo(ctx context.Context, in *ResolveAppInfoRequest, opts ...grpc.CallOption) (*ResolveAppInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ResolveAppInfoResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_ResolveAppInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) ReissueAppCert(ctx context.Context, in *ReissueAppCertRequest, opts ...grpc.CallOption) (*ReissueAppCertResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ReissueAppCertResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_ReissueAppCert_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) SignForApp(ctx context.Context, in *SignForAppRequest, opts ...grpc.CallOption) (*SignForAppResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SignForAppResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_SignForApp_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) OnNewConnection(ctx context.Context, in *OnNewConnectionRequest, opts ...grpc.CallOption) (*OnNewConnectionResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OnNewConnectionResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_OnNewConnection_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) OnInvalidLocalPort(ctx context.Context, in *OnInvalidLocalPortRequest, opts ...grpc.CallOption) (*OnInvalidLocalPortResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(OnInvalidLocalPortResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_OnInvalidLocalPort_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientApplicationServiceClient) GetTargetOSConfiguration(ctx context.Context, in *GetTargetOSConfigurationRequest, opts ...grpc.CallOption) (*GetTargetOSConfigurationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetTargetOSConfigurationResponse)
	err := c.cc.Invoke(ctx, ClientApplicationService_GetTargetOSConfiguration_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientApplicationServiceServer is the server API for ClientApplicationService service.
// All implementations must embed UnimplementedClientApplicationServiceServer
// for forward compatibility.
//
// ClientApplicationService is a service the VNet client applications provide to
// the VNet admin process to facilate app queries, certificate issuance,
// metrics, error reporting, and signatures.
type ClientApplicationServiceServer interface {
	// AuthenticateProcess mutually authenticates client applicates to the admin
	// service.
	AuthenticateProcess(context.Context, *AuthenticateProcessRequest) (*AuthenticateProcessResponse, error)
	// Ping is used by the admin process to regularly poll that the client
	// application is still running.
	Ping(context.Context, *PingRequest) (*PingResponse, error)
	// ResolveAppInfo returns info for the given app fqdn, or an error if the app
	// is not present in any logged-in cluster.
	ResolveAppInfo(context.Context, *ResolveAppInfoRequest) (*ResolveAppInfoResponse, error)
	// ReissueAppCert issues a new app cert.
	ReissueAppCert(context.Context, *ReissueAppCertRequest) (*ReissueAppCertResponse, error)
	// SignForApp issues a signature with the private key associated with an x509
	// certificate previously issued for a requested app.
	SignForApp(context.Context, *SignForAppRequest) (*SignForAppResponse, error)
	// OnNewConnection gets called whenever a new connection is about to be
	// established through VNet for observability.
	OnNewConnection(context.Context, *OnNewConnectionRequest) (*OnNewConnectionResponse, error)
	// OnInvalidLocalPort gets called before VNet refuses to handle a connection
	// to a multi-port TCP app because the provided port does not match any of the
	// TCP ports in the app spec.
	OnInvalidLocalPort(context.Context, *OnInvalidLocalPortRequest) (*OnInvalidLocalPortResponse, error)
	// GetTargetOSConfiguration gets the target OS configuration.
	GetTargetOSConfiguration(context.Context, *GetTargetOSConfigurationRequest) (*GetTargetOSConfigurationResponse, error)
	mustEmbedUnimplementedClientApplicationServiceServer()
}

// UnimplementedClientApplicationServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedClientApplicationServiceServer struct{}

func (UnimplementedClientApplicationServiceServer) AuthenticateProcess(context.Context, *AuthenticateProcessRequest) (*AuthenticateProcessResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuthenticateProcess not implemented")
}
func (UnimplementedClientApplicationServiceServer) Ping(context.Context, *PingRequest) (*PingResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (UnimplementedClientApplicationServiceServer) ResolveAppInfo(context.Context, *ResolveAppInfoRequest) (*ResolveAppInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ResolveAppInfo not implemented")
}
func (UnimplementedClientApplicationServiceServer) ReissueAppCert(context.Context, *ReissueAppCertRequest) (*ReissueAppCertResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReissueAppCert not implemented")
}
func (UnimplementedClientApplicationServiceServer) SignForApp(context.Context, *SignForAppRequest) (*SignForAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignForApp not implemented")
}
func (UnimplementedClientApplicationServiceServer) OnNewConnection(context.Context, *OnNewConnectionRequest) (*OnNewConnectionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnNewConnection not implemented")
}
func (UnimplementedClientApplicationServiceServer) OnInvalidLocalPort(context.Context, *OnInvalidLocalPortRequest) (*OnInvalidLocalPortResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method OnInvalidLocalPort not implemented")
}
func (UnimplementedClientApplicationServiceServer) GetTargetOSConfiguration(context.Context, *GetTargetOSConfigurationRequest) (*GetTargetOSConfigurationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTargetOSConfiguration not implemented")
}
func (UnimplementedClientApplicationServiceServer) mustEmbedUnimplementedClientApplicationServiceServer() {
}
func (UnimplementedClientApplicationServiceServer) testEmbeddedByValue() {}

// UnsafeClientApplicationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientApplicationServiceServer will
// result in compilation errors.
type UnsafeClientApplicationServiceServer interface {
	mustEmbedUnimplementedClientApplicationServiceServer()
}

func RegisterClientApplicationServiceServer(s grpc.ServiceRegistrar, srv ClientApplicationServiceServer) {
	// If the following call pancis, it indicates UnimplementedClientApplicationServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ClientApplicationService_ServiceDesc, srv)
}

func _ClientApplicationService_AuthenticateProcess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthenticateProcessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).AuthenticateProcess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_AuthenticateProcess_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).AuthenticateProcess(ctx, req.(*AuthenticateProcessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_Ping_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).Ping(ctx, req.(*PingRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_ResolveAppInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ResolveAppInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).ResolveAppInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_ResolveAppInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).ResolveAppInfo(ctx, req.(*ResolveAppInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_ReissueAppCert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReissueAppCertRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).ReissueAppCert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_ReissueAppCert_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).ReissueAppCert(ctx, req.(*ReissueAppCertRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_SignForApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignForAppRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).SignForApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_SignForApp_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).SignForApp(ctx, req.(*SignForAppRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_OnNewConnection_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnNewConnectionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).OnNewConnection(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_OnNewConnection_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).OnNewConnection(ctx, req.(*OnNewConnectionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_OnInvalidLocalPort_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OnInvalidLocalPortRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).OnInvalidLocalPort(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_OnInvalidLocalPort_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).OnInvalidLocalPort(ctx, req.(*OnInvalidLocalPortRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientApplicationService_GetTargetOSConfiguration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetTargetOSConfigurationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientApplicationServiceServer).GetTargetOSConfiguration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientApplicationService_GetTargetOSConfiguration_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientApplicationServiceServer).GetTargetOSConfiguration(ctx, req.(*GetTargetOSConfigurationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientApplicationService_ServiceDesc is the grpc.ServiceDesc for ClientApplicationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientApplicationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "teleport.lib.vnet.v1.ClientApplicationService",
	HandlerType: (*ClientApplicationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AuthenticateProcess",
			Handler:    _ClientApplicationService_AuthenticateProcess_Handler,
		},
		{
			MethodName: "Ping",
			Handler:    _ClientApplicationService_Ping_Handler,
		},
		{
			MethodName: "ResolveAppInfo",
			Handler:    _ClientApplicationService_ResolveAppInfo_Handler,
		},
		{
			MethodName: "ReissueAppCert",
			Handler:    _ClientApplicationService_ReissueAppCert_Handler,
		},
		{
			MethodName: "SignForApp",
			Handler:    _ClientApplicationService_SignForApp_Handler,
		},
		{
			MethodName: "OnNewConnection",
			Handler:    _ClientApplicationService_OnNewConnection_Handler,
		},
		{
			MethodName: "OnInvalidLocalPort",
			Handler:    _ClientApplicationService_OnInvalidLocalPort_Handler,
		},
		{
			MethodName: "GetTargetOSConfiguration",
			Handler:    _ClientApplicationService_GetTargetOSConfiguration_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "teleport/lib/vnet/v1/client_application_service.proto",
}
