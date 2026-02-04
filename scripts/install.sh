#!/bin/bash
#
# ServerHub Agent 一键安装脚本
#
# 使用方法:
#   curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/install.sh | sudo bash
#   或
#   curl -fsSL https://raw.githubusercontent.com/Zhang142857/serverhub/main/scripts/install.sh | sudo bash
#
# 环境变量:
#   SERVERHUB_VERSION  - 指定版本 (默认: latest)
#   SERVERHUB_TOKEN    - 预设认证令牌
#   SERVERHUB_PORT     - 监听端口 (默认: 9527)
#   INSTALL_DIR        - 安装目录 (默认: /usr/local/bin)
#   BUILD_FROM_SOURCE  - 设为 1 从源码构建 (需要 Go 环境)
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
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
BUILD_FROM_SOURCE="${BUILD_FROM_SOURCE:-0}"

# 包管理器
PKG_MANAGER=""
PKG_UPDATE=""
PKG_INSTALL=""

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

log_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# 检查是否为 root 用户
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "请使用 root 用户运行此脚本"
        log_info "尝试: sudo bash -c \"\$(curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/install.sh)\""
        exit 1
    fi
}

# 检测包管理器
detect_package_manager() {
    log_step "检测包管理器..."

    if command -v apt-get &> /dev/null; then
        PKG_MANAGER="apt"
        PKG_UPDATE="apt-get update -qq"
        PKG_INSTALL="apt-get install -y -qq"
        log_info "检测到 APT 包管理器 (Debian/Ubuntu)"
    elif command -v yum &> /dev/null; then
        PKG_MANAGER="yum"
        PKG_UPDATE="yum makecache -q"
        PKG_INSTALL="yum install -y -q"
        log_info "检测到 YUM 包管理器 (CentOS/RHEL)"
    elif command -v dnf &> /dev/null; then
        PKG_MANAGER="dnf"
        PKG_UPDATE="dnf makecache -q"
        PKG_INSTALL="dnf install -y -q"
        log_info "检测到 DNF 包管理器 (Fedora)"
    elif command -v pacman &> /dev/null; then
        PKG_MANAGER="pacman"
        PKG_UPDATE="pacman -Sy --noconfirm"
        PKG_INSTALL="pacman -S --noconfirm"
        log_info "检测到 Pacman 包管理器 (Arch Linux)"
    elif command -v apk &> /dev/null; then
        PKG_MANAGER="apk"
        PKG_UPDATE="apk update"
        PKG_INSTALL="apk add --no-cache"
        log_info "检测到 APK 包管理器 (Alpine)"
    elif command -v zypper &> /dev/null; then
        PKG_MANAGER="zypper"
        PKG_UPDATE="zypper refresh"
        PKG_INSTALL="zypper install -y"
        log_info "检测到 Zypper 包管理器 (openSUSE)"
    elif command -v brew &> /dev/null; then
        PKG_MANAGER="brew"
        PKG_UPDATE="brew update"
        PKG_INSTALL="brew install"
        log_info "检测到 Homebrew 包管理器 (macOS)"
    else
        log_warn "未检测到已知的包管理器，将跳过自动安装依赖"
        PKG_MANAGER="unknown"
    fi
}

# 安装单个包
install_package() {
    local pkg=$1
    local pkg_name=$pkg

    # 包名映射（不同发行版包名可能不同）
    case "$PKG_MANAGER" in
        apt)
            case "$pkg" in
                openssl) pkg_name="openssl" ;;
            esac
            ;;
        yum|dnf)
            case "$pkg" in
                openssl) pkg_name="openssl" ;;
            esac
            ;;
        pacman)
            case "$pkg" in
                openssl) pkg_name="openssl" ;;
            esac
            ;;
        apk)
            case "$pkg" in
                openssl) pkg_name="openssl" ;;
            esac
            ;;
    esac

    log_info "安装 $pkg_name..."
    if ! $PKG_INSTALL $pkg_name 2>/dev/null; then
        log_warn "安装 $pkg_name 失败，请手动安装"
        return 1
    fi
    return 0
}

# 检查并安装依赖
check_and_install_dependencies() {
    log_step "检查依赖..."

    local required_cmds=("curl" "tar")
    local optional_cmds=("openssl")
    local missing_required=()
    local missing_optional=()

    # 检查必需依赖
    for cmd in "${required_cmds[@]}"; do
        if ! command -v $cmd &> /dev/null; then
            missing_required+=($cmd)
        fi
    done

    # 检查可选依赖
    for cmd in "${optional_cmds[@]}"; do
        if ! command -v $cmd &> /dev/null; then
            missing_optional+=($cmd)
        fi
    done

    # 如果有缺失的依赖，尝试安装
    if [ ${#missing_required[@]} -ne 0 ] || [ ${#missing_optional[@]} -ne 0 ]; then
        if [ "$PKG_MANAGER" = "unknown" ]; then
            if [ ${#missing_required[@]} -ne 0 ]; then
                log_error "缺少必需依赖: ${missing_required[*]}"
                log_info "请手动安装后重试"
                exit 1
            fi
        else
            log_info "更新包索引..."
            $PKG_UPDATE 2>/dev/null || true

            # 安装必需依赖
            for pkg in "${missing_required[@]}"; do
                if ! install_package "$pkg"; then
                    log_error "无法安装必需依赖: $pkg"
                    exit 1
                fi
            done

            # 安装可选依赖
            for pkg in "${missing_optional[@]}"; do
                install_package "$pkg" || true
            done
        fi
    fi

    log_success "依赖检查完成"
}

# 检查网络连接
check_network() {
    log_step "检查网络连接..."

    # 尝试多个地址
    local test_urls=(
        "https://github.com"
        "https://api.github.com"
        "https://cdn.jsdelivr.net"
    )

    local connected=false
    for url in "${test_urls[@]}"; do
        if curl -fsSL --connect-timeout 5 "$url" &>/dev/null; then
            connected=true
            break
        fi
    done

    if [ "$connected" = false ]; then
        log_error "无法连接到网络，请检查网络设置"
        log_info "需要能够访问 GitHub 和 jsDelivr CDN"
        exit 1
    fi

    log_success "网络连接正常"
}

# 检查 systemd
check_systemd() {
    log_step "检查 systemd..."

    if ! command -v systemctl &> /dev/null; then
        log_warn "未检测到 systemd，将跳过服务配置"
        log_info "您需要手动配置服务管理"
        return 1
    fi

    if ! systemctl --version &>/dev/null; then
        log_warn "systemd 不可用，将跳过服务配置"
        return 1
    fi

    log_success "systemd 可用"
    return 0
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

# 检测 Linux 发行版
detect_distro() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        echo "$NAME $VERSION_ID"
    elif [ -f /etc/redhat-release ]; then
        cat /etc/redhat-release
    elif [ -f /etc/debian_version ]; then
        echo "Debian $(cat /etc/debian_version)"
    else
        echo "Unknown"
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

# 检查 Go 环境（用于源码构建）
check_go_environment() {
    log_step "检查 Go 环境..."

    if ! command -v go &> /dev/null; then
        log_warn "未检测到 Go 环境"

        if [ "$PKG_MANAGER" != "unknown" ]; then
            read -p "是否自动安装 Go? [y/N] " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                install_go
            else
                log_error "源码构建需要 Go 环境"
                exit 1
            fi
        else
            log_error "请先安装 Go 1.22+ 环境"
            log_info "下载地址: https://go.dev/dl/"
            exit 1
        fi
    fi

    # 检查 Go 版本
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local major=$(echo $go_version | cut -d. -f1)
    local minor=$(echo $go_version | cut -d. -f2)

    if [ "$major" -lt 1 ] || ([ "$major" -eq 1 ] && [ "$minor" -lt 22 ]); then
        log_error "Go 版本过低: $go_version，需要 1.22+"
        exit 1
    fi

    log_success "Go 环境正常: go$go_version"
}

# 安装 Go
install_go() {
    log_info "安装 Go..."

    local go_version="1.22.0"
    local arch=$(detect_arch)
    local os=$(detect_os)

    # 下载 Go
    local go_url="https://go.dev/dl/go${go_version}.${os}-${arch}.tar.gz"
    local tmp_dir=$(mktemp -d)

    log_info "下载 Go ${go_version}..."
    if ! curl -fsSL -o "${tmp_dir}/go.tar.gz" "$go_url"; then
        log_error "下载 Go 失败"
        rm -rf "$tmp_dir"
        exit 1
    fi

    # 安装
    log_info "安装 Go 到 /usr/local..."
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "${tmp_dir}/go.tar.gz"
    rm -rf "$tmp_dir"

    # 配置环境变量
    export PATH=$PATH:/usr/local/go/bin

    # 添加到 profile
    if ! grep -q '/usr/local/go/bin' /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi

    log_success "Go 安装完成"
}

# 从源码构建
build_from_source() {
    log_step "从源码构建..."

    check_go_environment

    local tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # 克隆仓库
    log_info "克隆仓库..."
    if ! command -v git &> /dev/null; then
        log_info "安装 git..."
        install_package "git"
    fi

    git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "${tmp_dir}/serverhub"

    # 构建
    log_info "编译 Agent..."
    cd "${tmp_dir}/serverhub/agent"
    CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o "${BINARY_NAME}" ./cmd/agent

    # 安装
    log_info "安装到 ${INSTALL_DIR}..."
    mkdir -p "${INSTALL_DIR}"
    mv "${BINARY_NAME}" "${INSTALL_DIR}/"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    log_success "源码构建完成"
}

# 下载并安装二进制文件
download_binary() {
    local os=$1
    local arch=$2
    local version=$3

    # 构建下载 URL
    local filename="${BINARY_NAME}_${os}_${arch}.tar.gz"
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${filename}"

    log_step "下载 ServerHub Agent ${version} (${os}/${arch})..."
    log_info "下载地址: ${download_url}"

    # 创建临时目录
    local tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # 下载文件
    if ! curl -fsSL -o "${tmp_dir}/${filename}" "${download_url}"; then
        log_error "下载失败"
        log_info "可能原因:"
        log_info "  1. 版本 ${version} 不存在"
        log_info "  2. 平台 ${os}/${arch} 不支持"
        log_info "  3. 网络连接问题"
        log_info ""
        log_info "可用版本请查看: https://github.com/${GITHUB_REPO}/releases"
        log_info ""

        # 询问是否从源码构建
        read -p "是否尝试从源码构建? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            BUILD_FROM_SOURCE=1
            build_from_source
            return
        fi
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

    log_step "创建配置文件..."
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
    log_step "创建 systemd 服务..."

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

# 创建 init.d 服务（非 systemd 系统）
create_initd_service() {
    log_step "创建 init.d 服务..."

    cat > "/etc/init.d/serverhub-agent" << 'EOF'
#!/bin/sh
### BEGIN INIT INFO
# Provides:          serverhub-agent
# Required-Start:    $network $remote_fs
# Required-Stop:     $network $remote_fs
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: ServerHub Agent
# Description:       ServerHub Agent - Server Management Agent
### END INIT INFO

DAEMON=/usr/local/bin/serverhub-agent
CONFIG=/etc/serverhub/agent.yaml
PIDFILE=/var/run/serverhub-agent.pid
NAME=serverhub-agent

case "$1" in
    start)
        echo "Starting $NAME..."
        start-stop-daemon --start --background --make-pidfile --pidfile $PIDFILE --exec $DAEMON -- -config $CONFIG
        ;;
    stop)
        echo "Stopping $NAME..."
        start-stop-daemon --stop --pidfile $PIDFILE
        rm -f $PIDFILE
        ;;
    restart)
        $0 stop
        sleep 1
        $0 start
        ;;
    status)
        if [ -f $PIDFILE ]; then
            if kill -0 $(cat $PIDFILE) 2>/dev/null; then
                echo "$NAME is running"
                exit 0
            fi
        fi
        echo "$NAME is not running"
        exit 1
        ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
exit 0
EOF

    chmod +x /etc/init.d/serverhub-agent

    # 添加到启动项
    if command -v update-rc.d &> /dev/null; then
        update-rc.d serverhub-agent defaults
    elif command -v chkconfig &> /dev/null; then
        chkconfig --add serverhub-agent
    fi

    log_success "init.d 服务已创建"
}

# 启动服务
start_service() {
    log_step "启动 ServerHub Agent..."

    if command -v systemctl &> /dev/null; then
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
    else
        /etc/init.d/serverhub-agent start
        sleep 2
        if /etc/init.d/serverhub-agent status &>/dev/null; then
            log_success "服务已启动"
        else
            log_error "服务启动失败"
            exit 1
        fi
    fi
}

# 显示安装信息
show_info() {
    local token=$1
    local ip=$(curl -fsSL --connect-timeout 5 ifconfig.me 2>/dev/null || curl -fsSL --connect-timeout 5 ipinfo.io/ip 2>/dev/null || echo "YOUR_SERVER_IP")

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
    echo "卸载: curl -fsSL https://cdn.jsdelivr.net/gh/Zhang142857/serverhub@main/scripts/uninstall.sh | sudo bash"
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

# 显示系统信息
show_system_info() {
    local os=$(detect_os)
    local arch=$(detect_arch)
    local distro=$(detect_distro)

    echo ""
    echo "========================================"
    echo "  系统信息"
    echo "========================================"
    echo "操作系统: ${os}"
    echo "架构: ${arch}"
    echo "发行版: ${distro}"
    echo "========================================"
    echo ""
}

# 主函数
main() {
    echo ""
    echo "========================================"
    echo "  ServerHub Agent 一键安装脚本"
    echo "========================================"
    echo ""

    # 检查 root
    check_root

    # 显示系统信息
    show_system_info

    # 检测包管理器
    detect_package_manager

    # 检查并安装依赖
    check_and_install_dependencies

    # 检查网络
    check_network

    # 检查已安装
    check_existing

    # 检测系统
    local os=$(detect_os)
    local arch=$(detect_arch)
    log_info "目标平台: ${os}/${arch}"

    # 获取版本
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
    fi
    log_info "安装版本: ${VERSION}"

    # 下载安装或从源码构建
    if [ "$BUILD_FROM_SOURCE" = "1" ]; then
        build_from_source
    else
        download_binary "$os" "$arch" "$VERSION"
    fi

    # 生成令牌
    local token=$(generate_token)

    # 创建配置
    create_config "$token"

    # 创建服务
    if check_systemd; then
        create_systemd_service
    else
        create_initd_service
    fi

    # 启动服务
    start_service

    # 显示信息
    show_info "$token"
}

# 运行
main "$@"
