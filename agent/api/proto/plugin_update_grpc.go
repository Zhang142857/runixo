// Hand-written gRPC service definitions for PluginService and UpdateService.
// These correspond to the proto definitions in proto/agent.proto.
package proto

import (
	"context"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// ==================== PluginService ====================

// PluginServiceClient is the client API for PluginService.
type PluginServiceClient interface {
	ListPlugins(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PluginList, error)
	InstallPlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*ActionResponse, error)
	UninstallPlugin(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*ActionResponse, error)
	EnablePlugin(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*ActionResponse, error)
	DisablePlugin(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*ActionResponse, error)
	GetPluginConfig(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*PluginConfig, error)
	SetPluginConfig(ctx context.Context, in *SetPluginConfigRequest, opts ...grpc.CallOption) (*ActionResponse, error)
	GetPluginStatus(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*PluginStatusMsg, error)
	GetAvailablePlugins(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*AvailablePluginList, error)
}

type pluginServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginServiceClient(cc grpc.ClientConnInterface) PluginServiceClient {
	return &pluginServiceClient{cc}
}

func (c *pluginServiceClient) ListPlugins(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*PluginList, error) {
	out := new(PluginList)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/ListPlugins", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) InstallPlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/InstallPlugin", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) UninstallPlugin(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/UninstallPlugin", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) EnablePlugin(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/EnablePlugin", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) DisablePlugin(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/DisablePlugin", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) GetPluginConfig(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*PluginConfig, error) {
	out := new(PluginConfig)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/GetPluginConfig", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) SetPluginConfig(ctx context.Context, in *SetPluginConfigRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/SetPluginConfig", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) GetPluginStatus(ctx context.Context, in *PluginRequest, opts ...grpc.CallOption) (*PluginStatusMsg, error) {
	out := new(PluginStatusMsg)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/GetPluginStatus", in, out, opts...)
	return out, err
}
func (c *pluginServiceClient) GetAvailablePlugins(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*AvailablePluginList, error) {
	out := new(AvailablePluginList)
	err := c.cc.Invoke(ctx, "/serverhub.PluginService/GetAvailablePlugins", in, out, opts...)
	return out, err
}

// PluginServiceServer is the server API for PluginService.
type PluginServiceServer interface {
	ListPlugins(context.Context, *Empty) (*PluginList, error)
	InstallPlugin(context.Context, *InstallPluginRequest) (*ActionResponse, error)
	UninstallPlugin(context.Context, *PluginRequest) (*ActionResponse, error)
	EnablePlugin(context.Context, *PluginRequest) (*ActionResponse, error)
	DisablePlugin(context.Context, *PluginRequest) (*ActionResponse, error)
	GetPluginConfig(context.Context, *PluginRequest) (*PluginConfig, error)
	SetPluginConfig(context.Context, *SetPluginConfigRequest) (*ActionResponse, error)
	GetPluginStatus(context.Context, *PluginRequest) (*PluginStatusMsg, error)
	GetAvailablePlugins(context.Context, *Empty) (*AvailablePluginList, error)
}

// UnimplementedPluginServiceServer should be embedded to have forward compatible implementations.
type UnimplementedPluginServiceServer struct{}

func (UnimplementedPluginServiceServer) ListPlugins(context.Context, *Empty) (*PluginList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListPlugins not implemented")
}
func (UnimplementedPluginServiceServer) InstallPlugin(context.Context, *InstallPluginRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InstallPlugin not implemented")
}
func (UnimplementedPluginServiceServer) UninstallPlugin(context.Context, *PluginRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UninstallPlugin not implemented")
}
func (UnimplementedPluginServiceServer) EnablePlugin(context.Context, *PluginRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EnablePlugin not implemented")
}
func (UnimplementedPluginServiceServer) DisablePlugin(context.Context, *PluginRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisablePlugin not implemented")
}
func (UnimplementedPluginServiceServer) GetPluginConfig(context.Context, *PluginRequest) (*PluginConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPluginConfig not implemented")
}
func (UnimplementedPluginServiceServer) SetPluginConfig(context.Context, *SetPluginConfigRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetPluginConfig not implemented")
}
func (UnimplementedPluginServiceServer) GetPluginStatus(context.Context, *PluginRequest) (*PluginStatusMsg, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetPluginStatus not implemented")
}
func (UnimplementedPluginServiceServer) GetAvailablePlugins(context.Context, *Empty) (*AvailablePluginList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAvailablePlugins not implemented")
}

// UnsafePluginServiceServer may be embedded to opt out of forward compatibility.
type UnsafePluginServiceServer interface {
	mustEmbedUnimplementedPluginServiceServer()
}

var _PluginService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "serverhub.PluginService",
	HandlerType: (*PluginServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "ListPlugins", Handler: _PluginService_ListPlugins_Handler},
		{MethodName: "InstallPlugin", Handler: _PluginService_InstallPlugin_Handler},
		{MethodName: "UninstallPlugin", Handler: _PluginService_UninstallPlugin_Handler},
		{MethodName: "EnablePlugin", Handler: _PluginService_EnablePlugin_Handler},
		{MethodName: "DisablePlugin", Handler: _PluginService_DisablePlugin_Handler},
		{MethodName: "GetPluginConfig", Handler: _PluginService_GetPluginConfig_Handler},
		{MethodName: "SetPluginConfig", Handler: _PluginService_SetPluginConfig_Handler},
		{MethodName: "GetPluginStatus", Handler: _PluginService_GetPluginStatus_Handler},
		{MethodName: "GetAvailablePlugins", Handler: _PluginService_GetAvailablePlugins_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/agent.proto",
}

func RegisterPluginServiceServer(s grpc.ServiceRegistrar, srv PluginServiceServer) {
	s.RegisterService(&_PluginService_serviceDesc, srv)
}

func _PluginService_ListPlugins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).ListPlugins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/ListPlugins"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).ListPlugins(ctx, req.(*Empty))
	})
}
func _PluginService_InstallPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).InstallPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/InstallPlugin"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).InstallPlugin(ctx, req.(*InstallPluginRequest))
	})
}
func _PluginService_UninstallPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).UninstallPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/UninstallPlugin"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).UninstallPlugin(ctx, req.(*PluginRequest))
	})
}
func _PluginService_EnablePlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).EnablePlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/EnablePlugin"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).EnablePlugin(ctx, req.(*PluginRequest))
	})
}
func _PluginService_DisablePlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).DisablePlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/DisablePlugin"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).DisablePlugin(ctx, req.(*PluginRequest))
	})
}
func _PluginService_GetPluginConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).GetPluginConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/GetPluginConfig"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).GetPluginConfig(ctx, req.(*PluginRequest))
	})
}
func _PluginService_SetPluginConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetPluginConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).SetPluginConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/SetPluginConfig"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).SetPluginConfig(ctx, req.(*SetPluginConfigRequest))
	})
}
func _PluginService_GetPluginStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).GetPluginStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/GetPluginStatus"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).GetPluginStatus(ctx, req.(*PluginRequest))
	})
}
func _PluginService_GetAvailablePlugins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServiceServer).GetAvailablePlugins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.PluginService/GetAvailablePlugins"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServiceServer).GetAvailablePlugins(ctx, req.(*Empty))
	})
}

// ==================== UpdateService ====================

// UpdateServiceServer is the server API for UpdateService.
type UpdateServiceServer interface {
	CheckUpdate(context.Context, *Empty) (*UpdateInfo, error)
	DownloadUpdate(*UpdateRequest, UpdateService_DownloadUpdateServer) error
	ApplyUpdate(context.Context, *UpdateRequest) (*ActionResponse, error)
	GetUpdateConfig(context.Context, *Empty) (*UpdateConfig, error)
	SetUpdateConfig(context.Context, *UpdateConfig) (*ActionResponse, error)
	GetUpdateHistory(context.Context, *Empty) (*UpdateHistory, error)
}

// UnimplementedUpdateServiceServer should be embedded to have forward compatible implementations.
type UnimplementedUpdateServiceServer struct{}

func (UnimplementedUpdateServiceServer) CheckUpdate(context.Context, *Empty) (*UpdateInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckUpdate not implemented")
}
func (UnimplementedUpdateServiceServer) DownloadUpdate(*UpdateRequest, UpdateService_DownloadUpdateServer) error {
	return status.Errorf(codes.Unimplemented, "method DownloadUpdate not implemented")
}
func (UnimplementedUpdateServiceServer) ApplyUpdate(context.Context, *UpdateRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApplyUpdate not implemented")
}
func (UnimplementedUpdateServiceServer) GetUpdateConfig(context.Context, *Empty) (*UpdateConfig, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUpdateConfig not implemented")
}
func (UnimplementedUpdateServiceServer) SetUpdateConfig(context.Context, *UpdateConfig) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetUpdateConfig not implemented")
}
func (UnimplementedUpdateServiceServer) GetUpdateHistory(context.Context, *Empty) (*UpdateHistory, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUpdateHistory not implemented")
}

// UnsafeUpdateServiceServer may be embedded to opt out of forward compatibility.
type UnsafeUpdateServiceServer interface {
	mustEmbedUnimplementedUpdateServiceServer()
}

// UpdateService_DownloadUpdateServer is the server-side streaming interface.
type UpdateService_DownloadUpdateServer interface {
	Send(*DownloadProgress) error
	grpc.ServerStream
}

type updateServiceDownloadUpdateServer struct {
	grpc.ServerStream
}

func (x *updateServiceDownloadUpdateServer) Send(m *DownloadProgress) error {
	return x.ServerStream.SendMsg(m)
}

// UpdateService_DownloadUpdateClient is the client-side streaming interface.
type UpdateService_DownloadUpdateClient interface {
	Recv() (*DownloadProgress, error)
	grpc.ClientStream
}

type updateServiceDownloadUpdateClient struct {
	grpc.ClientStream
}

func (x *updateServiceDownloadUpdateClient) Recv() (*DownloadProgress, error) {
	m := new(DownloadProgress)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _UpdateService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "serverhub.UpdateService",
	HandlerType: (*UpdateServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CheckUpdate", Handler: _UpdateService_CheckUpdate_Handler},
		{MethodName: "ApplyUpdate", Handler: _UpdateService_ApplyUpdate_Handler},
		{MethodName: "GetUpdateConfig", Handler: _UpdateService_GetUpdateConfig_Handler},
		{MethodName: "SetUpdateConfig", Handler: _UpdateService_SetUpdateConfig_Handler},
		{MethodName: "GetUpdateHistory", Handler: _UpdateService_GetUpdateHistory_Handler},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "DownloadUpdate",
			Handler:       _UpdateService_DownloadUpdate_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "proto/agent.proto",
}

func RegisterUpdateServiceServer(s grpc.ServiceRegistrar, srv UpdateServiceServer) {
	s.RegisterService(&_UpdateService_serviceDesc, srv)
}

func _UpdateService_CheckUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpdateServiceServer).CheckUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.UpdateService/CheckUpdate"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpdateServiceServer).CheckUpdate(ctx, req.(*Empty))
	})
}
func _UpdateService_DownloadUpdate_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(UpdateRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(UpdateServiceServer).DownloadUpdate(m, &updateServiceDownloadUpdateServer{stream})
}
func _UpdateService_ApplyUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpdateServiceServer).ApplyUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.UpdateService/ApplyUpdate"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpdateServiceServer).ApplyUpdate(ctx, req.(*UpdateRequest))
	})
}
func _UpdateService_GetUpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpdateServiceServer).GetUpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.UpdateService/GetUpdateConfig"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpdateServiceServer).GetUpdateConfig(ctx, req.(*Empty))
	})
}
func _UpdateService_SetUpdateConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConfig)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpdateServiceServer).SetUpdateConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.UpdateService/SetUpdateConfig"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpdateServiceServer).SetUpdateConfig(ctx, req.(*UpdateConfig))
	})
}
func _UpdateService_GetUpdateHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpdateServiceServer).GetUpdateHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/serverhub.UpdateService/GetUpdateHistory"}
	return interceptor(ctx, in, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpdateServiceServer).GetUpdateHistory(ctx, req.(*Empty))
	})
}
