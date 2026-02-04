package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Empty 空消息
type Empty struct{}

// 认证相关
type AuthRequest struct {
	Token         string
	ClientVersion string
}

type AuthResponse struct {
	Success      bool
	Message      string
	AgentVersion string
	ExpiresAt    int64
}

// 系统信息
type SystemInfo struct {
	Hostname        string
	Os              string
	Platform        string
	PlatformVersion string
	KernelVersion   string
	Arch            string
	Uptime          int64
	BootTime        int64
	Cpu             *CpuInfo
	Memory          *MemoryInfo
	Disks           []*DiskInfo
	Networks        []*NetworkInfo
	Gpus            []*GpuInfo
}

type CpuInfo struct {
	Model        string
	Cores        int32
	Threads      int32
	Frequency    float64
	UsagePerCore []float64
}

type MemoryInfo struct {
	Total       uint64
	Available   uint64
	Used        uint64
	UsedPercent float64
	SwapTotal   uint64
	SwapUsed    uint64
}

type DiskInfo struct {
	Device      string
	Mountpoint  string
	Fstype      string
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

type NetworkInfo struct {
	Name      string
	Addresses []string
	Mac       string
	BytesSent uint64
	BytesRecv uint64
}

type GpuInfo struct {
	Name          string
	DriverVersion string
	MemoryTotal   uint64
	MemoryUsed    uint64
	Temperature   float64
	Utilization   float64
}

// 监控指标
type MetricsRequest struct {
	IntervalSeconds int32
	Metrics         []string
}

type Metrics struct {
	Timestamp      *timestamppb.Timestamp
	CpuUsage       float64
	MemoryUsage    float64
	DiskMetrics    []*DiskMetric
	NetworkMetrics []*NetworkMetric
	Load1          float64
	Load5          float64
	Load15         float64
}

type DiskMetric struct {
	Device     string
	ReadBytes  uint64
	WriteBytes uint64
	ReadCount  uint64
	WriteCount uint64
}

type NetworkMetric struct {
	Interface   string
	BytesSent   uint64
	BytesRecv   uint64
	PacketsSent uint64
	PacketsRecv uint64
}

// 命令执行
type CommandRequest struct {
	Command        string
	Args           []string
	WorkingDir     string
	Env            map[string]string
	TimeoutSeconds int32
	Sudo           bool
}

type CommandResponse struct {
	ExitCode   int32
	Stdout     string
	Stderr     string
	DurationMs int64
}

type ShellInput struct {
	Input isShellInput_Input
}

type isShellInput_Input interface {
	isShellInput_Input()
}

type ShellInput_Start struct {
	Start *ShellStart
}

type ShellInput_Data struct {
	Data []byte
}

type ShellInput_Resize struct {
	Resize *ShellResize
}

func (*ShellInput_Start) isShellInput_Input()  {}
func (*ShellInput_Data) isShellInput_Input()   {}
func (*ShellInput_Resize) isShellInput_Input() {}

func (x *ShellInput) GetStart() *ShellStart {
	if x, ok := x.Input.(*ShellInput_Start); ok {
		return x.Start
	}
	return nil
}

func (x *ShellInput) GetData() []byte {
	if x, ok := x.Input.(*ShellInput_Data); ok {
		return x.Data
	}
	return nil
}

func (x *ShellInput) GetResize() *ShellResize {
	if x, ok := x.Input.(*ShellInput_Resize); ok {
		return x.Resize
	}
	return nil
}

type ShellStart struct {
	Shell string
	Rows  int32
	Cols  int32
	Env   map[string]string
}

type ShellResize struct {
	Rows int32
	Cols int32
}

type ShellOutput struct {
	Data []byte
}

// 文件操作
type FileRequest struct {
	Path string
}

type FileContent struct {
	Content []byte
	Info    *FileInfo
}

type FileInfo struct {
	Name    string
	Path    string
	Size    int64
	Mode    int64
	ModTime int64
	IsDir   bool
	Owner   string
	Group   string
}

type WriteFileRequest struct {
	Path       string
	Content    []byte
	Mode       int64
	CreateDirs bool
}

type DirRequest struct {
	Path       string
	Recursive  bool
	ShowHidden bool
}

type DirContent struct {
	Path  string
	Files []*FileInfo
}

type FileChunk struct {
	Path   string
	Data   []byte
	Offset int64
	IsLast bool
}

// 日志
type LogRequest struct {
	Path   string
	Lines  int32
	Follow bool
}

type LogLine struct {
	Content   string
	Timestamp *timestamppb.Timestamp
}

// 服务管理
type ServiceFilter struct {
	NameFilter   string
	StatusFilter string
}

type ServiceList struct {
	Services []*ServiceInfo
}

type ServiceInfo struct {
	Name        string
	Status      string
	Description string
	Enabled     bool
	Pid         int32
	Uptime      int64
}

type ServiceActionRequest struct {
	Name   string
	Action ServiceAction
}

type ServiceAction int32

const (
	ServiceAction_SERVICE_START   ServiceAction = 0
	ServiceAction_SERVICE_STOP    ServiceAction = 1
	ServiceAction_SERVICE_RESTART ServiceAction = 2
	ServiceAction_SERVICE_ENABLE  ServiceAction = 3
	ServiceAction_SERVICE_DISABLE ServiceAction = 4
)

// 进程管理
type ProcessFilter struct {
	NameFilter string
	UserFilter string
}

type ProcessList struct {
	Processes []*ProcessInfo
}

type ProcessInfo struct {
	Pid           int32
	Ppid          int32
	Name          string
	User          string
	Status        string
	CpuPercent    float64
	MemoryPercent float64
	MemoryRss     uint64
	CreateTime    int64
	Cmdline       string
}

type KillProcessRequest struct {
	Pid    int32
	Signal int32
}

// 通用响应
type ActionResponse struct {
	Success bool
	Message string
	Error   string
}

// 流式服务接口
type AgentService_GetMetricsServer interface {
	Send(*Metrics) error
	grpc.ServerStream
}

type AgentService_ExecuteShellServer interface {
	Send(*ShellOutput) error
	Recv() (*ShellInput, error)
	grpc.ServerStream
}

type AgentService_TailLogServer interface {
	Send(*LogLine) error
	grpc.ServerStream
}

// UnimplementedAgentServiceServer 用于向前兼容
type UnimplementedAgentServiceServer struct{}

func (UnimplementedAgentServiceServer) Authenticate(context.Context, *AuthRequest) (*AuthResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) GetSystemInfo(context.Context, *Empty) (*SystemInfo, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) GetMetrics(*MetricsRequest, AgentService_GetMetricsServer) error {
	return nil
}
func (UnimplementedAgentServiceServer) ExecuteCommand(context.Context, *CommandRequest) (*CommandResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) ExecuteShell(AgentService_ExecuteShellServer) error {
	return nil
}
func (UnimplementedAgentServiceServer) ReadFile(context.Context, *FileRequest) (*FileContent, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) WriteFile(context.Context, *WriteFileRequest) (*ActionResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) ListDirectory(context.Context, *DirRequest) (*DirContent, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) DeleteFile(context.Context, *FileRequest) (*ActionResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) TailLog(*LogRequest, AgentService_TailLogServer) error {
	return nil
}
func (UnimplementedAgentServiceServer) ListServices(context.Context, *ServiceFilter) (*ServiceList, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) ServiceAction(context.Context, *ServiceActionRequest) (*ActionResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) ListProcesses(context.Context, *ProcessFilter) (*ProcessList, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) KillProcess(context.Context, *KillProcessRequest) (*ActionResponse, error) {
	return nil, nil
}
func (UnimplementedAgentServiceServer) mustEmbedUnimplementedAgentServiceServer() {}
