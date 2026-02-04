package server

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/creack/pty"
	"github.com/rs/zerolog/log"
	"github.com/serverhub/agent/internal/collector"
	"github.com/serverhub/agent/internal/executor"
	"github.com/serverhub/agent/internal/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// 全局路径验证器
var pathValidator = security.NewPathValidator(security.DefaultSecurityConfig())

// AgentServiceServer 定义服务接口
type AgentServiceServer interface {
	Authenticate(context.Context, *AuthRequest) (*AuthResponse, error)
	GetSystemInfo(context.Context, *Empty) (*SystemInfo, error)
	GetMetrics(*MetricsRequest, AgentService_GetMetricsServer) error
	ExecuteCommand(context.Context, *CommandRequest) (*CommandResponse, error)
	ExecuteShell(AgentService_ExecuteShellServer) error
	ReadFile(context.Context, *FileRequest) (*FileContent, error)
	WriteFile(context.Context, *WriteFileRequest) (*ActionResponse, error)
	ListDirectory(context.Context, *DirRequest) (*DirContent, error)
	DeleteFile(context.Context, *FileRequest) (*ActionResponse, error)
	TailLog(*LogRequest, AgentService_TailLogServer) error
	ListServices(context.Context, *ServiceFilter) (*ServiceList, error)
	ServiceAction(context.Context, *ServiceActionRequest) (*ActionResponse, error)
	ListProcesses(context.Context, *ProcessFilter) (*ProcessList, error)
	KillProcess(context.Context, *KillProcessRequest) (*ActionResponse, error)
	mustEmbedUnimplementedAgentServiceServer()
}

// AgentServer 实现 AgentServiceServer
type AgentServer struct {
	UnimplementedAgentServiceServer
	version   string
	collector *collector.Collector
}

// NewAgentServer 创建新的 AgentServer
func NewAgentServer(version string) *AgentServer {
	return &AgentServer{
		version:   version,
		collector: collector.New(),
	}
}

// Authenticate 认证
func (s *AgentServer) Authenticate(ctx context.Context, req *AuthRequest) (*AuthResponse, error) {
	return &AuthResponse{
		Success:      true,
		Message:      "认证成功",
		AgentVersion: s.version,
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(),
	}, nil
}

// GetSystemInfo 获取系统信息
func (s *AgentServer) GetSystemInfo(ctx context.Context, req *Empty) (*SystemInfo, error) {
	info, err := s.collector.GetSystemInfo()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "获取系统信息失败: %v", err)
	}
	return convertSystemInfo(info), nil
}

// GetMetrics 获取实时监控指标流
func (s *AgentServer) GetMetrics(req *MetricsRequest, stream AgentService_GetMetricsServer) error {
	interval := time.Duration(req.IntervalSeconds) * time.Second
	if interval < time.Second {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-ticker.C:
			metrics, err := s.collector.GetMetrics()
			if err != nil {
				log.Error().Err(err).Msg("采集指标失败")
				continue
			}
			serverMetrics := convertMetrics(metrics)
			serverMetrics.Timestamp = timestamppb.Now()
			if err := stream.Send(serverMetrics); err != nil {
				return err
			}
		}
	}
}

// ExecuteCommand 执行命令
func (s *AgentServer) ExecuteCommand(ctx context.Context, req *CommandRequest) (*CommandResponse, error) {
	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	result, err := executor.Execute(ctx, req.Command, req.Args, executor.Options{
		WorkingDir: req.WorkingDir,
		Env:        req.Env,
		Timeout:    timeout,
		Sudo:       req.Sudo,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "执行命令失败: %v", err)
	}

	return &CommandResponse{
		ExitCode:   int32(result.ExitCode),
		Stdout:     result.Stdout,
		Stderr:     result.Stderr,
		DurationMs: result.DurationMs,
	}, nil
}

// ExecuteShell 交互式 Shell
func (s *AgentServer) ExecuteShell(stream AgentService_ExecuteShellServer) error {
	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}

	start := firstMsg.GetStart()
	if start == nil {
		return status.Error(codes.InvalidArgument, "首条消息必须是启动消息")
	}

	shell := start.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
		if shell == "" {
			shell = "/bin/bash"
		}
	}

	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	for k, v := range start.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(start.Rows),
		Cols: uint16(start.Cols),
	})
	if err != nil {
		return status.Errorf(codes.Internal, "启动 PTY 失败: %v", err)
	}
	defer ptmx.Close()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Error().Err(err).Msg("读取 PTY 失败")
				}
				return
			}
			if err := stream.Send(&ShellOutput{Data: buf[:n]}); err != nil {
				return
			}
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		switch input := msg.Input.(type) {
		case *ShellInput_Data:
			if _, err := ptmx.Write(input.Data); err != nil {
				return status.Errorf(codes.Internal, "写入 PTY 失败: %v", err)
			}
		case *ShellInput_Resize:
			if err := pty.Setsize(ptmx, &pty.Winsize{
				Rows: uint16(input.Resize.Rows),
				Cols: uint16(input.Resize.Cols),
			}); err != nil {
				log.Error().Err(err).Msg("调整 PTY 大小失败")
			}
		}
	}
}

// ReadFile 读取文件
func (s *AgentServer) ReadFile(ctx context.Context, req *FileRequest) (*FileContent, error) {
	content, info, err := executor.ReadFile(req.Path)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "读取文件失败: %v", err)
	}
	return &FileContent{
		Content: content,
		Info:    convertFileInfo(info),
	}, nil
}

// WriteFile 写入文件
func (s *AgentServer) WriteFile(ctx context.Context, req *WriteFileRequest) (*ActionResponse, error) {
	if err := executor.WriteFile(req.Path, req.Content, req.Mode, req.CreateDirs); err != nil {
		return &ActionResponse{Success: false, Error: err.Error()}, nil
	}
	return &ActionResponse{Success: true, Message: "文件已保存"}, nil
}

// ListDirectory 列出目录
func (s *AgentServer) ListDirectory(ctx context.Context, req *DirRequest) (*DirContent, error) {
	files, err := executor.ListDirectory(req.Path, req.Recursive, req.ShowHidden)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "列出目录失败: %v", err)
	}
	return &DirContent{
		Path:  req.Path,
		Files: convertFileInfoList(files),
	}, nil
}

// DeleteFile 删除文件（带安全检查）
func (s *AgentServer) DeleteFile(ctx context.Context, req *FileRequest) (*ActionResponse, error) {
	cleanPath, err := security.SanitizePath(req.Path)
	if err != nil {
		return &ActionResponse{Success: false, Error: "路径安全检查失败: " + err.Error()}, nil
	}

	if err := pathValidator.ValidatePathForWrite(cleanPath); err != nil {
		return &ActionResponse{Success: false, Error: "删除路径被拒绝: " + err.Error()}, nil
	}

	realPath, err := filepath.EvalSymlinks(cleanPath)
	if err == nil && realPath != cleanPath {
		if err := pathValidator.ValidatePathForWrite(realPath); err != nil {
			return &ActionResponse{Success: false, Error: "符号链接目标路径被拒绝: " + err.Error()}, nil
		}
	}

	forbiddenPaths := []string{"/", "/bin", "/sbin", "/usr", "/etc", "/var", "/boot", "/root", "/home"}
	for _, forbidden := range forbiddenPaths {
		if cleanPath == forbidden {
			return &ActionResponse{Success: false, Error: "禁止删除系统关键目录"}, nil
		}
	}

	if err := os.RemoveAll(cleanPath); err != nil {
		return &ActionResponse{Success: false, Error: err.Error()}, nil
	}
	return &ActionResponse{Success: true, Message: "文件已删除"}, nil
}

// TailLog 日志流
func (s *AgentServer) TailLog(req *LogRequest, stream AgentService_TailLogServer) error {
	lineChan, err := executor.TailFile(stream.Context(), req.Path, int(req.Lines), req.Follow)
	if err != nil {
		return status.Errorf(codes.Internal, "读取日志失败: %v", err)
	}

	for line := range lineChan {
		if err := stream.Send(&LogLine{
			Content:   line,
			Timestamp: timestamppb.Now(),
		}); err != nil {
			return err
		}
	}
	return nil
}

// ListServices 列出服务
func (s *AgentServer) ListServices(ctx context.Context, req *ServiceFilter) (*ServiceList, error) {
	services, err := executor.ListServices(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出服务失败: %v", err)
	}
	return &ServiceList{Services: convertServiceList(services)}, nil
}

// ServiceAction 服务操作
func (s *AgentServer) ServiceAction(ctx context.Context, req *ServiceActionRequest) (*ActionResponse, error) {
	var action string
	switch req.Action {
	case ServiceAction_SERVICE_START:
		action = "start"
	case ServiceAction_SERVICE_STOP:
		action = "stop"
	case ServiceAction_SERVICE_RESTART:
		action = "restart"
	case ServiceAction_SERVICE_ENABLE:
		action = "enable"
	case ServiceAction_SERVICE_DISABLE:
		action = "disable"
	default:
		return nil, status.Error(codes.InvalidArgument, "未知操作")
	}

	if err := executor.ServiceAction(ctx, req.Name, action); err != nil {
		return &ActionResponse{Success: false, Error: err.Error()}, nil
	}
	return &ActionResponse{Success: true, Message: "操作成功"}, nil
}

// ListProcesses 列出进程
func (s *AgentServer) ListProcesses(ctx context.Context, req *ProcessFilter) (*ProcessList, error) {
	processes, err := s.collector.ListProcesses()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出进程失败: %v", err)
	}
	return &ProcessList{Processes: convertProcessList(processes)}, nil
}

// KillProcess 终止进程
func (s *AgentServer) KillProcess(ctx context.Context, req *KillProcessRequest) (*ActionResponse, error) {
	if err := executor.KillProcess(int(req.Pid), int(req.Signal)); err != nil {
		return &ActionResponse{Success: false, Error: err.Error()}, nil
	}
	return &ActionResponse{Success: true, Message: "进程已终止"}, nil
}

// RegisterAgentServiceServer 注册服务到 gRPC 服务器
func RegisterAgentServiceServer(s *grpc.Server, srv AgentServiceServer) {
	s.RegisterService(&AgentService_ServiceDesc, srv)
}

// AgentService_ServiceDesc gRPC 服务描述
var AgentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "serverhub.AgentService",
	HandlerType: (*AgentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _AgentService_Authenticate_Handler,
		},
		{
			MethodName: "GetSystemInfo",
			Handler:    _AgentService_GetSystemInfo_Handler,
		},
		{
			MethodName: "ExecuteCommand",
			Handler:    _AgentService_ExecuteCommand_Handler,
		},
		{
			MethodName: "ReadFile",
			Handler:    _AgentService_ReadFile_Handler,
		},
		{
			MethodName: "WriteFile",
			Handler:    _AgentService_WriteFile_Handler,
		},
		{
			MethodName: "ListDirectory",
			Handler:    _AgentService_ListDirectory_Handler,
		},
		{
			MethodName: "DeleteFile",
			Handler:    _AgentService_DeleteFile_Handler,
		},
		{
			MethodName: "ListServices",
			Handler:    _AgentService_ListServices_Handler,
		},
		{
			MethodName: "ServiceAction",
			Handler:    _AgentService_ServiceAction_Handler,
		},
		{
			MethodName: "ListProcesses",
			Handler:    _AgentService_ListProcesses_Handler,
		},
		{
			MethodName: "KillProcess",
			Handler:    _AgentService_KillProcess_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetMetrics",
			Handler:       _AgentService_GetMetrics_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "ExecuteShell",
			Handler:       _AgentService_ExecuteShell_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "TailLog",
			Handler:       _AgentService_TailLog_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "agent.proto",
}

// Handler 函数
func _AgentService_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/Authenticate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).Authenticate(ctx, req.(*AuthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_GetSystemInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).GetSystemInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/GetSystemInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).GetSystemInfo(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_ExecuteCommand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).ExecuteCommand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/ExecuteCommand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).ExecuteCommand(ctx, req.(*CommandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_ReadFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).ReadFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/ReadFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).ReadFile(ctx, req.(*FileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_WriteFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).WriteFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/WriteFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).WriteFile(ctx, req.(*WriteFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_ListDirectory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DirRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).ListDirectory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/ListDirectory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).ListDirectory(ctx, req.(*DirRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_DeleteFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).DeleteFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/DeleteFile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).DeleteFile(ctx, req.(*FileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_ListServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).ListServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/ListServices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).ListServices(ctx, req.(*ServiceFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_ServiceAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ServiceActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).ServiceAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/ServiceAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).ServiceAction(ctx, req.(*ServiceActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_ListProcesses_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProcessFilter)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).ListProcesses(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/ListProcesses",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).ListProcesses(ctx, req.(*ProcessFilter))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_KillProcess_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KillProcessRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentServiceServer).KillProcess(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/serverhub.AgentService/KillProcess",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentServiceServer).KillProcess(ctx, req.(*KillProcessRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentService_GetMetrics_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(MetricsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AgentServiceServer).GetMetrics(m, &agentServiceGetMetricsServer{stream})
}

type agentServiceGetMetricsServer struct {
	grpc.ServerStream
}

func (x *agentServiceGetMetricsServer) Send(m *Metrics) error {
	return x.ServerStream.SendMsg(m)
}

func _AgentService_ExecuteShell_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(AgentServiceServer).ExecuteShell(&agentServiceExecuteShellServer{stream})
}

type agentServiceExecuteShellServer struct {
	grpc.ServerStream
}

func (x *agentServiceExecuteShellServer) Send(m *ShellOutput) error {
	return x.ServerStream.SendMsg(m)
}

func (x *agentServiceExecuteShellServer) Recv() (*ShellInput, error) {
	m := new(ShellInput)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _AgentService_TailLog_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LogRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(AgentServiceServer).TailLog(m, &agentServiceTailLogServer{stream})
}

type agentServiceTailLogServer struct {
	grpc.ServerStream
}

func (x *agentServiceTailLogServer) Send(m *LogLine) error {
	return x.ServerStream.SendMsg(m)
}
