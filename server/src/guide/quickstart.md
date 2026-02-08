# 快速开始

## 方式一：SSH 自动安装（推荐）

1. 打开客户端，进入「服务器」页面
2. 点击「SSH 安装」按钮
3. 填写服务器 SSH 连接信息
4. 等待自动安装完成

安装脚本会自动下载 Agent、生成 TLS 证书、配置 systemd 服务并将证书回传给客户端。

## 方式二：手动安装

### 安装 Agent

```bash
# 一键安装
curl -fsSL https://raw.githubusercontent.com/Zhang142857/runixo/main/scripts/install.sh | sudo bash

# 或手动下载
wget https://github.com/Zhang142857/runixo/releases/latest/download/runixo-agent_linux_amd64.tar.gz
tar -xzf runixo-agent_linux_amd64.tar.gz
sudo mv runixo-agent /usr/local/bin/
```

### 查看连接信息

```bash
sudo runixo info        # 查看 IP、端口和 Token
sudo runixo token       # 仅查看 Token
```

### 在客户端添加服务器

1. 打开客户端 →「服务器」→「添加服务器」
2. 填写 IP、端口、Token
3. 如连接失败，点击服务器操作菜单 →「导入证书」
4. 粘贴 `sudo cat /var/lib/runixo/tls/cert.pem` 的输出

## 安装客户端

从 [GitHub Releases](https://github.com/Zhang142857/runixo/releases) 下载：

| 平台 | 文件 |
|------|------|
| Windows | `Runixo-Setup-x.x.x.exe` |
| macOS | `Runixo-x.x.x.dmg` |
| Linux | `Runixo-x.x.x.AppImage` |
