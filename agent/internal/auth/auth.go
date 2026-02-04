package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// 安全配置
const (
	MaxFailedAttempts    = 5               // 最大失败尝试次数
	LockoutDuration      = 15 * time.Minute // 锁定时间
	TokenMinLength       = 32              // 令牌最小长度
)

// AuthInterceptor 认证拦截器
type AuthInterceptor struct {
	token         string
	requireAuth   bool
	failedAttempts map[string]*attemptInfo
	mu            sync.RWMutex
}

type attemptInfo struct {
	count     int
	lockedUntil time.Time
}

// NewAuthInterceptor 创建认证拦截器
func NewAuthInterceptor(token string) *AuthInterceptor {
	// 强制要求认证令牌
	requireAuth := token != ""
	if !requireAuth {
		// 如果没有配置令牌，生成一个随机令牌并记录警告
		// 在生产环境中应该强制配置令牌
		token, _ = GenerateToken()
	}

	return &AuthInterceptor{
		token:         token,
		requireAuth:   requireAuth,
		failedAttempts: make(map[string]*attemptInfo),
	}
}

// IsAuthRequired 返回是否需要认证
func (a *AuthInterceptor) IsAuthRequired() bool {
	return a.requireAuth
}

// GetToken 获取当前令牌（仅用于显示生成的令牌）
func (a *AuthInterceptor) GetToken() string {
	return a.token
}

// Unary 一元调用拦截器
func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 跳过认证方法本身
		if info.FullMethod == "/serverhub.AgentService/Authenticate" {
			return handler(ctx, req)
		}

		if err := a.authorize(ctx); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// Stream 流式调用拦截器
func (a *AuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if err := a.authorize(ss.Context()); err != nil {
			return err
		}
		return handler(srv, ss)
	}
}

// getClientIP 获取客户端 IP
func (a *AuthInterceptor) getClientIP(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "unknown"
}

// isLocked 检查 IP 是否被锁定
func (a *AuthInterceptor) isLocked(ip string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if info, exists := a.failedAttempts[ip]; exists {
		if time.Now().Before(info.lockedUntil) {
			return true
		}
	}
	return false
}

// recordFailedAttempt 记录失败尝试
func (a *AuthInterceptor) recordFailedAttempt(ip string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.failedAttempts[ip]; !exists {
		a.failedAttempts[ip] = &attemptInfo{}
	}

	info := a.failedAttempts[ip]
	info.count++

	if info.count >= MaxFailedAttempts {
		info.lockedUntil = time.Now().Add(LockoutDuration)
		return true // 已锁定
	}
	return false
}

// resetFailedAttempts 重置失败尝试
func (a *AuthInterceptor) resetFailedAttempts(ip string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.failedAttempts, ip)
}

// authorize 验证请求
func (a *AuthInterceptor) authorize(ctx context.Context) error {
	clientIP := a.getClientIP(ctx)

	// 检查是否被锁定
	if a.isLocked(clientIP) {
		return status.Error(codes.ResourceExhausted, "认证失败次数过多，请稍后重试")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		a.recordFailedAttempt(clientIP)
		return status.Error(codes.Unauthenticated, "缺少元数据")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		a.recordFailedAttempt(clientIP)
		return status.Error(codes.Unauthenticated, "缺少认证令牌")
	}

	token := values[0]
	// 支持 Bearer token 格式
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	// 使用常量时间比较防止时序攻击
	if subtle.ConstantTimeCompare([]byte(token), []byte(a.token)) != 1 {
		locked := a.recordFailedAttempt(clientIP)
		if locked {
			return status.Error(codes.ResourceExhausted, "认证失败次数过多，账户已锁定")
		}
		return status.Error(codes.Unauthenticated, "认证令牌无效")
	}

	// 认证成功，重置失败计数
	a.resetFailedAttempts(clientIP)
	return nil
}

// GenerateToken 生成随机令牌
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateToken 验证令牌格式
func ValidateToken(token string) bool {
	return len(token) >= TokenMinLength
}
