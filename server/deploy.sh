#!/bin/bash
# Runixo 官方服务器一键部署脚本
# 用法: TUNNEL_TOKEN=xxx ./deploy.sh
#
# 首次使用需要先在 Cloudflare Dashboard 创建 Tunnel:
#   1. 进入 Cloudflare Zero Trust → Networks → Tunnels
#   2. 创建 Tunnel，记下 Token
#   3. 设置 Public Hostname: api.runixo.dev → http://localhost:8080

set -e

# --- 配置 ---
APP_DIR="/opt/runixo-server"
SERVICE_NAME="runixo-server"

if [ -z "$TUNNEL_TOKEN" ]; then
    echo "错误: 请设置 TUNNEL_TOKEN 环境变量"
    echo "用法: TUNNEL_TOKEN=xxx ./deploy.sh"
    exit 1
fi

echo "=== Runixo Server 部署开始 ==="

# 1. 安装依赖
echo "[1/5] 安装依赖..."
apt-get update -qq && apt-get install -y -qq curl sqlite3 > /dev/null

# 2. 安装 cloudflared
if ! command -v cloudflared &> /dev/null; then
    echo "[2/5] 安装 cloudflared..."
    curl -fsSL https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o /usr/local/bin/cloudflared
    chmod +x /usr/local/bin/cloudflared
else
    echo "[2/5] cloudflared 已安装"
fi

# 3. 部署应用
echo "[3/5] 部署应用..."
mkdir -p ${APP_DIR}/data
# 如果二进制不存在，提示构建
if [ ! -f "${APP_DIR}/runixo-server" ]; then
    echo "  请将构建好的 runixo-server 二进制放到 ${APP_DIR}/"
    echo "  构建命令: cd server && make build && cp runixo-server ${APP_DIR}/"
fi

# 写入配置（如果不存在）
if [ ! -f "${APP_DIR}/config.yaml" ]; then
    cat > ${APP_DIR}/config.yaml << 'EOF'
server:
  port: 8080
  trusted_proxies: ["127.0.0.1", "172.16.0.0/12"]
  read_timeout_seconds: 10
  write_timeout_seconds: 10
database:
  path: "data/runixo.db"
rate_limit:
  requests_per_minute: 60
  burst_size: 10
  ban_threshold: 300
  ban_duration_minutes: 30
github:
  owner: "runixo"
  repo: "runixo"
cdn:
  base_url: ""
EOF
fi

# 4. 创建 systemd 服务
echo "[4/5] 配置 systemd 服务..."

cat > /etc/systemd/system/${SERVICE_NAME}.service << EOF
[Unit]
Description=Runixo Official Server
After=network.target

[Service]
Type=simple
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/runixo-server config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

cat > /etc/systemd/system/cloudflared-tunnel.service << EOF
[Unit]
Description=Cloudflare Tunnel for Runixo
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/usr/local/bin/cloudflared tunnel run --token ${TUNNEL_TOKEN}
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 5. 启动服务
echo "[5/5] 启动服务..."
systemctl daemon-reload
systemctl enable ${SERVICE_NAME} cloudflared-tunnel
systemctl restart ${SERVICE_NAME} cloudflared-tunnel

echo ""
echo "=== 部署完成 ==="
echo "  应用状态: systemctl status ${SERVICE_NAME}"
echo "  隧道状态: systemctl status cloudflared-tunnel"
echo "  查看日志: journalctl -u ${SERVICE_NAME} -f"
echo ""
echo "迁移到新服务器时，只需在新机器上重新运行此脚本即可，无需修改 DNS。"
