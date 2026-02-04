#!/bin/bash
#
# ServerHub Agent 卸载脚本
#
# 使用方法:
#   curl -fsSL https://cdn.jsdelivr.net/gh/serverhub/serverhub@main/scripts/uninstall.sh | bash
#

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
BINARY_NAME="serverhub-agent"
CONFIG_DIR="/etc/serverhub"
SERVICE_FILE="/etc/systemd/system/serverhub-agent.service"
INSTALL_DIR="/usr/local/bin"

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
        exit 1
    fi
}

# 停止并禁用服务
stop_service() {
    if systemctl is-active --quiet serverhub-agent 2>/dev/null; then
        log_info "停止服务..."
        systemctl stop serverhub-agent
    fi

    if systemctl is-enabled --quiet serverhub-agent 2>/dev/null; then
        log_info "禁用服务..."
        systemctl disable serverhub-agent
    fi
}

# 删除文件
remove_files() {
    # 删除服务文件
    if [ -f "${SERVICE_FILE}" ]; then
        log_info "删除 systemd 服务文件..."
        rm -f "${SERVICE_FILE}"
        systemctl daemon-reload
    fi

    # 删除二进制文件
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        log_info "删除二进制文件..."
        rm -f "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # 询问是否删除配置文件
    if [ -d "${CONFIG_DIR}" ]; then
        read -p "是否删除配置文件 (${CONFIG_DIR})? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "删除配置目录..."
            rm -rf "${CONFIG_DIR}"
        else
            log_info "保留配置文件"
        fi
    fi
}

# 主函数
main() {
    echo ""
    echo "========================================"
    echo "  ServerHub Agent 卸载脚本"
    echo "========================================"
    echo ""

    check_root

    read -p "确定要卸载 ServerHub Agent? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "取消卸载"
        exit 0
    fi

    stop_service
    remove_files

    echo ""
    log_success "ServerHub Agent 已卸载"
    echo ""
}

main "$@"
