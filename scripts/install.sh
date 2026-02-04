#!/bin/bash
#
# ServerHub Agent 一键安装脚本
#
# 使用方法:
#   curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh | bash
#   或
#   curl -fsSL https://raw.githubusercontent.com/serverhub/serverhub/main/scripts/install.sh | bash
#
# 环境变量:
#   SERVERHUB_VERSION  - 指定版本 (默认: latest)
#   SERVERHUB_TOKEN    - 预设认证令牌
#   SERVERHUB_PORT     - 监听端口 (默认: 9527)
#   INSTALL_DIR        - 安装目录 (默认: /usr/local/bin)
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
GITHUB_REPO="Zhang142857/serverhub"
BINARY_NAME="serverhub-agent"
CONFIG_DIR="/etc/serverhub"
CONFIG_FILE="${CONFIG_DIR}/agent.yaml"
SERVICE_FILE="/etc/systemd/system/serverhub-agent.service"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${SERVERHUB_VERSION:-latest}"
PORT="${SERVERHUB_PORT:-9527}"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        log_info "尝试: sudo bash -c \"\$(curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/install.sh)\""
        exit 1
    fi
}

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        armv7l|armv7)
            echo "armv7"
            ;;
        i386|i686)
            echo "386"
            ;;
        *)
            log_error "不支持的架构: $arch"
            exit 1
            ;;
    esac
}

# 检测操作系统
detect_os() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case $os in
        linux)
            echo "linux"
            ;;
        darwin)
            echo "darwin"
            ;;
        freebsd)
            echo "freebsd"
            ;;
        *)
            log_error "不支持的操作系统: $os"
            exit 1
            ;;
    esac
}

# 检查依赖
check_dependencies() {
    local missing=()

    for cmd in curl tar; do
        if ! command -v $cmd &> /dev/null; then
            missing+=($cmd)
        fi
    done

    if [ ${#missing[@]} -ne 0 ]; then
        log_error "缺少依赖: ${missing[*]}"
        log_info "请先安装: apt-get install -y ${missing[*]} 或 yum install -y ${missing[*]}"
        exit 1
    fi
}

# 获取最新版本号
get_latest_version() {
    local latest=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$latest" ]; then
        log_warn "无法获取最新版本，使用默认版本 v0.1.0"
        echo "v0.1.0"
    else
        echo "$latest"
    fi
}

# 下载并安装二进制文件
download_binary() {
    local os=$1
    local arch=$2
    local version=$3

    # 构建下载 URL
    local filename="${BINARY_NAME}_${os}_${arch}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${filename}"

    log_info "下载 ServerHub Agent ${version} (${os}/${arch})..."
    log_info "下载地址: ${download_url}"

    # 创建临时目录
    local tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # 下载文件
    if ! curl -fsSL -o "${tmp_dir}/${filename}" "${download_url}"; then
        log_error "下载失败，请检查网络连接或版本号是否正确"
        log_info "可用版本请查看: https://github.com/${GITHUB_REPO}/releases"
        exit 1
    fi

    # 解压
    log_info "解压文件..."
    tar -xzf "${tmp_dir}/${filename}" -C "${tmp_dir}"

    # 安装二进制文件
    log_info "安装到 ${INSTALL_DIR}..."
    mkdir -p "${INSTALL_DIR}"
    mv "${tmp_dir}/${BINARY_NAME}" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    log_success "二进制文件安装完成"
}

# 生成认证令牌
generate_token() {
    if [ -n "${SERVERHUB_TOKEN}" ]; then
        echo "${SERVERHUB_TOKEN}"
    else
        # 使用 openssl 或 /dev/urandom 生成随机令牌
        if command -v openssl &> /dev/null; then
            openssl rand -hex 32
        else
            head -c 32 /dev/urandom | xxd -p | tr -d '\n'
        fi
    fi
}

# 创建配置文件
create_config() {
    local token=$1

    log_info "创建配置文件..."
    mkdir -p "${CONFIG_DIR}"

    cat > "${CONFIG_FILE}" << EOF
# ServerHub Agent 配置文件
# 由安装脚本自动生成

server:
  host: "0.0.0.0"
  port: ${PORT}
  tls:
    enabled: false
    cert: "${CONFIG_DIR}/cert.pem"
    key: "${CONFIG_DIR}/key.pem"

auth:
  token: "${token}"

metrics:
  interval: 2

log:
  level: "info"
EOF

    # 设置安全权限
    chmod 600 "${CONFIG_FILE}"

    log_success "配置文件已创建: ${CONFIG_FILE}"
}

# 创建 systemd 服务
create_systemd_service() {
    log_info "创建 systemd 服务..."

    cat > "${SERVICE_FILE}" << EOF
[Unit]
Description=ServerHub Agent - 服务器管理代理
Documentation=https://github.com/${GITHUB_REPO}
After=network.target

[Service]
Type=simple
User=root
ExecStart=${INSTALL_DIR}/${BINARY_NAME} -config ${CONFIG_FILE}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# 安全加固
NoNewPrivileges=false
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/var/log /tmp
PrivateTmp=true

[Install]
WantedBy=multi-user.target
EOF

    # 重新加载 systemd
    systemctl daemon-reload

    log_success "systemd 服务已创建"
}

# 启动服务
start_service() {
    log_info "启动 ServerHub Agent..."

    systemctl enable serverhub-agent
    systemctl start serverhub-agent

    # 等待服务启动
    sleep 2

    if systemctl is-active --quiet serverhub-agent; then
        log_success "服务已启动"
    else
        log_error "服务启动失败，请检查日志: journalctl -u serverhub-agent -f"
        exit 1
    fi
}

# 显示安装信息
show_info() {
    local token=$1
    local ip=$(curl -fsSL ifconfig.me 2>/dev/null || echo "YOUR_SERVER_IP")

    echo ""
    echo "========================================"
    echo -e "${GREEN}ServerHub Agent 安装成功!${NC}"
    echo "========================================"
    echo ""
    echo "服务状态: systemctl status serverhub-agent"
    echo "查看日志: journalctl -u serverhub-agent -f"
    echo "配置文件: ${CONFIG_FILE}"
    echo ""
    echo "========================================"
    echo -e "${YELLOW}连接信息 (请保存)${NC}"
    echo "========================================"
    echo "服务器地址: ${ip}"
    echo "端口: ${PORT}"
    echo -e "认证令牌: ${GREEN}${token}${NC}"
    echo ""
    echo "在 ServerHub 客户端中添加服务器时使用以上信息"
    echo ""
    echo "========================================"
    echo "常用命令"
    echo "========================================"
    echo "重启服务: systemctl restart serverhub-agent"
    echo "停止服务: systemctl stop serverhub-agent"
    echo "卸载: curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/uninstall.sh | bash"
    echo ""
}

# 检查是否已安装
check_existing() {
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        log_warn "检测到已安装的 ServerHub Agent"
        read -p "是否覆盖安装? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "取消安装"
            exit 0
        fi

        # 停止现有服务
        if systemctl is-active --quiet serverhub-agent 2>/dev/null; then
            log_info "停止现有服务..."
            systemctl stop serverhub-agent
        fi
    fi
}

# 主函数
main() {
    echo ""
    echo "========================================"
    echo "  ServerHub Agent 一键安装脚本"
    echo "========================================"
    echo ""

    # 检查
    check_root
    check_dependencies
    check_existing

    # 检测系统
    local os=$(detect_os)
    local arch=$(detect_arch)
    log_info "检测到系统: ${os}/${arch}"

    # 获取版本
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
    fi
    log_info "安装版本: ${VERSION}"

    # 下载安装
    download_binary "$os" "$arch" "$VERSION"

    # 生成令牌
    local token=$(generate_token)

    # 创建配置
    create_config "$token"

    # 创建服务
    create_systemd_service

    # 启动服务
    start_service

    # 显示信息
    show_info "$token"
}

# 运行
main "$@"
