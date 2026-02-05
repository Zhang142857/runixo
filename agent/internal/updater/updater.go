// Package updater Agent 自动更新系统
package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Config 更新配置
type Config struct {
	AutoUpdate    bool   `json:"auto_update"`
	CheckInterval int    `json:"check_interval"` // 秒
	UpdateChannel string `json:"update_channel"` // stable, beta, nightly
	LastCheck     string `json:"last_check"`
	NotifyOnly    bool   `json:"notify_only"` // 仅通知，不自动安装
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		AutoUpdate:    false,
		CheckInterval: 3600, // 1小时
		UpdateChannel: "stable",
		NotifyOnly:    true,
	}
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	Available      bool   `json:"available"`
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	ReleaseNotes   string `json:"release_notes"`
	DownloadURL    string `json:"download_url"`
	Size           int64  `json:"size"`
	Checksum       string `json:"checksum"`
	ReleaseDate    string `json:"release_date"`
	IsCritical     bool   `json:"is_critical"`
}

// UpdateRecord 更新记录
type UpdateRecord struct {
	Version     string `json:"version"`
	FromVersion string `json:"from_version"`
	Timestamp   int64  `json:"timestamp"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Percent    int    `json:"percent"`
	Status     string `json:"status"` // downloading, verifying, ready
}

// Updater 更新器
type Updater struct {
	config         *Config
	currentVersion string
	dataDir        string
	updateURL      string
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	checkTicker    *time.Ticker
	history        []UpdateRecord
	progressChan   chan *DownloadProgress
}

// NewUpdater 创建更新器
func NewUpdater(currentVersion, dataDir string) (*Updater, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	u := &Updater{
		config:         DefaultConfig(),
		currentVersion: currentVersion,
		dataDir:        dataDir,
		updateURL:      "https://releases.serverhub.dev",
		ctx:            ctx,
		cancel:         cancel,
		progressChan:   make(chan *DownloadProgress, 10),
	}

	// 加载配置
	u.loadConfig()
	u.loadHistory()

	return u, nil
}

// loadConfig 加载配置
func (u *Updater) loadConfig() {
	configFile := filepath.Join(u.dataDir, "update_config.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warn().Err(err).Msg("加载更新配置失败")
		}
		return
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Warn().Err(err).Msg("解析更新配置失败")
		return
	}

	u.config = &config
}

// saveConfig 保存配置
func (u *Updater) saveConfig() error {
	configFile := filepath.Join(u.dataDir, "update_config.json")
	data, err := json.MarshalIndent(u.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFile, data, 0644)
}

// loadHistory 加载更新历史
func (u *Updater) loadHistory() {
	historyFile := filepath.Join(u.dataDir, "update_history.json")
	data, err := os.ReadFile(historyFile)
	if err != nil {
		return
	}

	json.Unmarshal(data, &u.history)
}

// saveHistory 保存更新历史
func (u *Updater) saveHistory() error {
	historyFile := filepath.Join(u.dataDir, "update_history.json")
	data, err := json.MarshalIndent(u.history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(historyFile, data, 0644)
}

// Start 启动更新器
func (u *Updater) Start() {
	if !u.config.AutoUpdate {
		log.Info().Msg("自动更新已禁用")
		return
	}

	interval := time.Duration(u.config.CheckInterval) * time.Second
	u.checkTicker = time.NewTicker(interval)

	go func() {
		// 启动时检查一次
		u.checkAndUpdate()

		for {
			select {
			case <-u.ctx.Done():
				return
			case <-u.checkTicker.C:
				u.checkAndUpdate()
			}
		}
	}()

	log.Info().
		Int("interval", u.config.CheckInterval).
		Str("channel", u.config.UpdateChannel).
		Msg("自动更新已启动")
}

// Stop 停止更新器
func (u *Updater) Stop() {
	u.cancel()
	if u.checkTicker != nil {
		u.checkTicker.Stop()
	}
}

// checkAndUpdate 检查并更新
func (u *Updater) checkAndUpdate() {
	info, err := u.CheckUpdate()
	if err != nil {
		log.Warn().Err(err).Msg("检查更新失败")
		return
	}

	if !info.Available {
		log.Debug().Msg("当前已是最新版本")
		return
	}

	log.Info().
		Str("current", info.CurrentVersion).
		Str("latest", info.LatestVersion).
		Bool("critical", info.IsCritical).
		Msg("发现新版本")

	if u.config.NotifyOnly && !info.IsCritical {
		log.Info().Msg("仅通知模式，跳过自动更新")
		return
	}

	// 下载并应用更新
	if err := u.DownloadAndApply(info); err != nil {
		log.Error().Err(err).Msg("更新失败")
		u.recordUpdate(info.LatestVersion, false, err.Error())
	}
}

// CheckUpdate 检查更新
func (u *Updater) CheckUpdate() (*UpdateInfo, error) {
	u.mu.Lock()
	u.config.LastCheck = time.Now().Format(time.RFC3339)
	u.saveConfig()
	u.mu.Unlock()

	// 构建检查 URL
	url := fmt.Sprintf("%s/api/check?version=%s&channel=%s&os=%s&arch=%s",
		u.updateURL,
		u.currentVersion,
		u.config.UpdateChannel,
		runtime.GOOS,
		runtime.GOARCH,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求更新服务器失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("更新服务器返回错误: %s", resp.Status)
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("解析更新信息失败: %w", err)
	}

	info.CurrentVersion = u.currentVersion
	return &info, nil
}

// DownloadUpdate 下载更新
func (u *Updater) DownloadUpdate(version string, progressChan chan<- *DownloadProgress) (string, error) {
	// 获取更新信息
	info, err := u.CheckUpdate()
	if err != nil {
		return "", err
	}

	if !info.Available || info.LatestVersion != version {
		return "", fmt.Errorf("版本 %s 不可用", version)
	}

	// 创建下载目录
	downloadDir := filepath.Join(u.dataDir, "downloads")
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", err
	}

	// 下载文件
	downloadPath := filepath.Join(downloadDir, fmt.Sprintf("serverhub-agent-%s", version))
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	if err := u.downloadFile(info.DownloadURL, downloadPath, info.Size, progressChan); err != nil {
		return "", err
	}

	// 验证校验和
	if progressChan != nil {
		progressChan <- &DownloadProgress{
			Downloaded: info.Size,
			Total:      info.Size,
			Percent:    100,
			Status:     "verifying",
		}
	}

	if info.Checksum != "" {
		valid, err := verifyChecksum(downloadPath, info.Checksum)
		if err != nil {
			return "", fmt.Errorf("验证校验和失败: %w", err)
		}
		if !valid {
			os.Remove(downloadPath)
			return "", fmt.Errorf("校验和不匹配")
		}
	}

	if progressChan != nil {
		progressChan <- &DownloadProgress{
			Downloaded: info.Size,
			Total:      info.Size,
			Percent:    100,
			Status:     "ready",
		}
	}

	return downloadPath, nil
}

// downloadFile 下载文件
func (u *Updater) downloadFile(url, destPath string, totalSize int64, progressChan chan<- *DownloadProgress) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 带进度的复制
	var downloaded int64
	buf := make([]byte, 32*1024)

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			if progressChan != nil && totalSize > 0 {
				percent := int(float64(downloaded) / float64(totalSize) * 100)
				progressChan <- &DownloadProgress{
					Downloaded: downloaded,
					Total:      totalSize,
					Percent:    percent,
					Status:     "downloading",
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// ApplyUpdate 应用更新
func (u *Updater) ApplyUpdate(version string) error {
	downloadDir := filepath.Join(u.dataDir, "downloads")
	downloadPath := filepath.Join(downloadDir, fmt.Sprintf("serverhub-agent-%s", version))
	if runtime.GOOS == "windows" {
		downloadPath += ".exe"
	}

	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		return fmt.Errorf("更新文件不存在: %s", downloadPath)
	}

	// 获取当前可执行文件路径
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前可执行文件路径失败: %w", err)
	}

	// 备份当前版本
	backupPath := currentExe + ".backup"
	if err := os.Rename(currentExe, backupPath); err != nil {
		return fmt.Errorf("备份当前版本失败: %w", err)
	}

	// 复制新版本
	if err := copyFile(downloadPath, currentExe); err != nil {
		// 恢复备份
		os.Rename(backupPath, currentExe)
		return fmt.Errorf("安装新版本失败: %w", err)
	}

	// 设置执行权限
	if runtime.GOOS != "windows" {
		if err := os.Chmod(currentExe, 0755); err != nil {
			log.Warn().Err(err).Msg("设置执行权限失败")
		}
	}

	// 记录更新
	u.recordUpdate(version, true, "")

	// 清理
	os.Remove(downloadPath)
	os.Remove(backupPath)

	log.Info().Str("version", version).Msg("更新已应用，需要重启服务")

	// 重启服务
	go u.restartService()

	return nil
}

// DownloadAndApply 下载并应用更新
func (u *Updater) DownloadAndApply(info *UpdateInfo) error {
	progressChan := make(chan *DownloadProgress, 10)
	defer close(progressChan)

	// 启动进度日志
	go func() {
		for p := range progressChan {
			log.Debug().
				Int64("downloaded", p.Downloaded).
				Int64("total", p.Total).
				Int("percent", p.Percent).
				Str("status", p.Status).
				Msg("下载进度")
		}
	}()

	// 下载
	_, err := u.DownloadUpdate(info.LatestVersion, progressChan)
	if err != nil {
		return err
	}

	// 应用
	return u.ApplyUpdate(info.LatestVersion)
}

// restartService 重启服务
func (u *Updater) restartService() {
	time.Sleep(2 * time.Second)

	// 尝试使用 systemctl 重启
	if runtime.GOOS == "linux" {
		cmd := exec.Command("systemctl", "restart", "serverhub-agent")
		if err := cmd.Run(); err != nil {
			log.Warn().Err(err).Msg("systemctl 重启失败，尝试直接重启")
		} else {
			return
		}
	}

	// 直接退出，让进程管理器重启
	log.Info().Msg("正在重启...")
	os.Exit(0)
}

// recordUpdate 记录更新
func (u *Updater) recordUpdate(version string, success bool, errMsg string) {
	record := UpdateRecord{
		Version:     version,
		FromVersion: u.currentVersion,
		Timestamp:   time.Now().Unix(),
		Success:     success,
		Error:       errMsg,
	}

	u.mu.Lock()
	u.history = append(u.history, record)
	// 只保留最近 50 条记录
	if len(u.history) > 50 {
		u.history = u.history[len(u.history)-50:]
	}
	u.saveHistory()
	u.mu.Unlock()
}

// GetConfig 获取配置
func (u *Updater) GetConfig() *Config {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.config
}

// SetConfig 设置配置
func (u *Updater) SetConfig(config *Config) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.config = config

	// 重新启动定时检查
	if u.checkTicker != nil {
		u.checkTicker.Stop()
	}

	if config.AutoUpdate {
		interval := time.Duration(config.CheckInterval) * time.Second
		u.checkTicker = time.NewTicker(interval)
	}

	return u.saveConfig()
}

// GetHistory 获取更新历史
func (u *Updater) GetHistory() []UpdateRecord {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.history
}

// GetCurrentVersion 获取当前版本
func (u *Updater) GetCurrentVersion() string {
	return u.currentVersion
}

// verifyChecksum 验证校验和
func verifyChecksum(filePath, expected string) (bool, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	return actual == expected, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
