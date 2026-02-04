# ServerHub 安装脚本

本目录包含 ServerHub Agent 的安装和构建脚本。

## 一键安装

在目标服务器上执行以下命令即可完成安装：

```bash
# 使用 jsDelivr CDN (推荐，国内访问快)
curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh | sudo bash

# 使用 GitHub Raw (备用)
curl -fsSL https://raw.githubusercontent.com/serverhub/serverhub/main/scripts/install.sh | sudo bash
```

### 自定义安装

通过环境变量自定义安装：

```bash
# 指定版本
SERVERHUB_VERSION=v0.1.0 curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh | sudo bash

# 指定端口
SERVERHUB_PORT=9528 curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh | sudo bash

# 预设认证令牌
SERVERHUB_TOKEN=your_token_here curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh | sudo bash

# 组合使用
SERVERHUB_VERSION=v0.1.0 SERVERHUB_PORT=9528 curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh | sudo bash
```

## 卸载

```bash
curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/uninstall.sh | sudo bash
```

## 脚本说明

| 脚本 | 说明 |
|------|------|
| `install.sh` | 一键安装脚本，自动检测系统架构并安装 |
| `uninstall.sh` | 卸载脚本，移除 Agent 及相关文件 |
| `build.sh` | 跨平台构建脚本，生成所有平台的二进制文件 |

## 支持的平台

| 操作系统 | 架构 | 说明 |
|----------|------|------|
| Linux | x86_64 (amd64) | 大多数云服务器 |
| Linux | ARM64 (aarch64) | 树莓派 4、AWS Graviton |
| Linux | ARMv7 | 树莓派 3、旧版 ARM 设备 |
| Linux | x86 (386) | 32位系统 |
| macOS | x86_64 | Intel Mac |
| macOS | ARM64 | Apple Silicon (M1/M2/M3) |
| FreeBSD | x86_64 | FreeBSD 服务器 |

## 安装后

安装完成后，脚本会显示连接信息：

```
========================================
连接信息 (请保存)
========================================
服务器地址: xxx.xxx.xxx.xxx
端口: 9527
认证令牌: xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

在 ServerHub 客户端中使用这些信息添加服务器。

## 常用命令

```bash
# 查看服务状态
systemctl status serverhub-agent

# 查看日志
journalctl -u serverhub-agent -f

# 重启服务
systemctl restart serverhub-agent

# 停止服务
systemctl stop serverhub-agent

# 重新生成令牌
serverhub-agent --gen-token
```

## 配置文件

配置文件位于 `/etc/serverhub/agent.yaml`：

```yaml
server:
  host: "0.0.0.0"
  port: 9527
  tls:
    enabled: false
    cert: "/etc/serverhub/cert.pem"
    key: "/etc/serverhub/key.pem"

auth:
  token: "your_token_here"

metrics:
  interval: 2

log:
  level: "info"
```

## 启用 TLS

1. 准备证书文件：
   ```bash
   sudo cp your_cert.pem /etc/serverhub/cert.pem
   sudo cp your_key.pem /etc/serverhub/key.pem
   sudo chmod 600 /etc/serverhub/*.pem
   ```

2. 修改配置文件：
   ```yaml
   server:
     tls:
       enabled: true
       cert: "/etc/serverhub/cert.pem"
       key: "/etc/serverhub/key.pem"
   ```

3. 重启服务：
   ```bash
   sudo systemctl restart serverhub-agent
   ```

## 防火墙配置

确保防火墙允许 Agent 端口（默认 9527）：

```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 9527/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=9527/tcp
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 9527 -j ACCEPT
```

## 故障排除

### 安装失败

1. 检查网络连接
2. 确保使用 root 用户或 sudo
3. 检查系统架构是否支持

### 服务无法启动

```bash
# 查看详细错误
journalctl -u serverhub-agent -n 50

# 手动运行查看错误
/usr/local/bin/serverhub-agent -config /etc/serverhub/agent.yaml
```

### 无法连接

1. 检查防火墙设置
2. 确认端口未被占用
3. 验证认证令牌是否正确
