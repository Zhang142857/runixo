# ServerHub Agent 安装脚本

## 一键安装

```bash
curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/install.sh | sudo bash
```

安装完成后会显示服务器连接信息（公网IP、端口、认证令牌），可直接在客户端使用。

### 自定义安装

```bash
# 指定端口
SERVERHUB_PORT=8080 curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/install.sh | sudo bash

# 预设令牌
SERVERHUB_TOKEN=your-secret-token curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/install.sh | sudo bash
```

## serverhub 管理命令

安装后可使用 `serverhub` 命令管理 Agent：

```bash
serverhub info          # 查看连接信息（公网IP、端口、令牌）
serverhub status        # 查看服务状态
serverhub start         # 启动服务
serverhub stop          # 停止服务
serverhub restart       # 重启服务
serverhub logs          # 查看日志
serverhub token         # 显示当前令牌
serverhub token:reset   # 重置认证令牌
serverhub port <端口>   # 修改监听端口
serverhub config        # 编辑配置文件
serverhub uninstall     # 卸载 Agent
serverhub help          # 查看帮助
```

## 卸载

```bash
# 方式一：使用 serverhub 命令
sudo serverhub uninstall

# 方式二：使用卸载脚本
curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/uninstall.sh | sudo bash
```

## 支持的平台

| 操作系统 | 架构 |
|----------|------|
| Linux | x86_64 (amd64) |
| Linux | ARM64 (aarch64) |
| Linux | ARMv7 |
| macOS | x86_64 / ARM64 |

## 配置文件

位于 `/etc/serverhub/agent.yaml`，可通过 `serverhub config` 编辑。

## 故障排除

```bash
# 查看服务状态
serverhub status

# 查看日志
serverhub logs

# 手动运行查看错误
/usr/local/bin/serverhub-agent -config /etc/serverhub/agent.yaml
```
