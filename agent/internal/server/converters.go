package server

import (
	"context"

	"github.com/serverhub/agent/internal/collector"
	"github.com/serverhub/agent/internal/executor"
)

// convertSystemInfo 转换系统信息
func convertSystemInfo(info *collector.SystemInfo) *SystemInfo {
	if info == nil {
		return nil
	}

	result := &SystemInfo{
		Hostname:        info.Hostname,
		Os:              info.Os,
		Platform:        info.Platform,
		PlatformVersion: info.PlatformVersion,
		KernelVersion:   info.KernelVersion,
		Arch:            info.Arch,
		Uptime:          info.Uptime,
		BootTime:        info.BootTime,
	}

	if info.Cpu != nil {
		result.Cpu = &CpuInfo{
			Model:        info.Cpu.Model,
			Cores:        info.Cpu.Cores,
			Threads:      info.Cpu.Threads,
			Frequency:    info.Cpu.Frequency,
			UsagePerCore: info.Cpu.UsagePerCore,
		}
	}

	if info.Memory != nil {
		result.Memory = &MemoryInfo{
			Total:       info.Memory.Total,
			Available:   info.Memory.Available,
			Used:        info.Memory.Used,
			UsedPercent: info.Memory.UsedPercent,
			SwapTotal:   info.Memory.SwapTotal,
			SwapUsed:    info.Memory.SwapUsed,
		}
	}

	for _, d := range info.Disks {
		result.Disks = append(result.Disks, &DiskInfo{
			Device:      d.Device,
			Mountpoint:  d.Mountpoint,
			Fstype:      d.Fstype,
			Total:       d.Total,
			Used:        d.Used,
			Free:        d.Free,
			UsedPercent: d.UsedPercent,
		})
	}

	for _, n := range info.Networks {
		result.Networks = append(result.Networks, &NetworkInfo{
			Name:      n.Name,
			Addresses: n.Addresses,
			Mac:       n.Mac,
			BytesSent: n.BytesSent,
			BytesRecv: n.BytesRecv,
		})
	}

	return result
}

// convertMetrics 转换监控指标
func convertMetrics(m *collector.Metrics) *Metrics {
	if m == nil {
		return nil
	}

	result := &Metrics{
		CpuUsage:    m.CpuUsage,
		MemoryUsage: m.MemoryUsage,
		Load1:       m.Load1,
		Load5:       m.Load5,
		Load15:      m.Load15,
	}

	for _, d := range m.DiskMetrics {
		result.DiskMetrics = append(result.DiskMetrics, &DiskMetric{
			Device:     d.Device,
			ReadBytes:  d.ReadBytes,
			WriteBytes: d.WriteBytes,
			ReadCount:  d.ReadCount,
			WriteCount: d.WriteCount,
		})
	}

	for _, n := range m.NetworkMetrics {
		result.NetworkMetrics = append(result.NetworkMetrics, &NetworkMetric{
			Interface:   n.Interface,
			BytesSent:   n.BytesSent,
			BytesRecv:   n.BytesRecv,
			PacketsSent: n.PacketsSent,
			PacketsRecv: n.PacketsRecv,
		})
	}

	return result
}

// convertFileInfo 转换文件信息
func convertFileInfo(f *executor.FileInfo) *FileInfo {
	if f == nil {
		return nil
	}
	return &FileInfo{
		Name:    f.Name,
		Path:    f.Path,
		Size:    f.Size,
		Mode:    f.Mode,
		ModTime: f.ModTime,
		IsDir:   f.IsDir,
		Owner:   f.Owner,
		Group:   f.Group,
	}
}

// convertFileInfoList 转换文件信息列表
func convertFileInfoList(files []*executor.FileInfo) []*FileInfo {
	var result []*FileInfo
	for _, f := range files {
		result = append(result, convertFileInfo(f))
	}
	return result
}

// convertServiceList 转换服务列表
func convertServiceList(services []*executor.ServiceInfo) []*ServiceInfo {
	var result []*ServiceInfo
	for _, s := range services {
		result = append(result, &ServiceInfo{
			Name:        s.Name,
			Status:      s.Status,
			Description: s.Description,
			Enabled:     s.Enabled,
			Pid:         s.Pid,
			Uptime:      s.Uptime,
		})
	}
	return result
}

// convertProcessList 转换进程列表
func convertProcessList(processes []*collector.ProcessInfo) []*ProcessInfo {
	var result []*ProcessInfo
	for _, p := range processes {
		result = append(result, &ProcessInfo{
			Pid:           p.Pid,
			Ppid:          p.Ppid,
			Name:          p.Name,
			User:          p.User,
			Status:        p.Status,
			CpuPercent:    p.CpuPercent,
			MemoryPercent: p.MemoryPercent,
			MemoryRss:     p.MemoryRss,
			CreateTime:    p.CreateTime,
			Cmdline:       p.Cmdline,
		})
	}
	return result
}

// 流式服务接口实现
type metricsServerStream struct {
	ctx    context.Context
	sendFn func(*Metrics) error
}

func (s *metricsServerStream) Send(m *Metrics) error {
	return s.sendFn(m)
}

func (s *metricsServerStream) Context() context.Context {
	return s.ctx
}

type shellServerStream struct {
	ctx    context.Context
	sendFn func(*ShellOutput) error
	recvFn func() (*ShellInput, error)
}

func (s *shellServerStream) Send(o *ShellOutput) error {
	return s.sendFn(o)
}

func (s *shellServerStream) Recv() (*ShellInput, error) {
	return s.recvFn()
}

func (s *shellServerStream) Context() context.Context {
	return s.ctx
}

type tailLogServerStream struct {
	ctx    context.Context
	sendFn func(*LogLine) error
}

func (s *tailLogServerStream) Send(l *LogLine) error {
	return s.sendFn(l)
}

func (s *tailLogServerStream) Context() context.Context {
	return s.ctx
}
