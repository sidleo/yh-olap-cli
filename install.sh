#!/bin/bash
set -e

# yh-olap-cli 安装脚本 (macOS/Linux)

REPO="sidleo/yh-olap-cli"
BINARY_NAME="yh-olap-cli"
INSTALL_DIR="/usr/local/bin"
SKILL_DIR="$HOME/.agents/skills/yh-olap"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 检测操作系统和架构
detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*|MSYS*|CYGWIN*)  os="windows";;
        *)          error "不支持的操作系统: $(uname -s)"
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)              error "不支持的架构: $(uname -m)"
    esac

    echo "${os}_${arch}"
}

# 下载最新版本
download_latest() {
    local platform=$1
    local version

    info "获取最新版本..."
    version=$(gh api repos/${REPO}/releases/latest --jq '.tag_name' 2>/dev/null || \
              curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$version" ]; then
        error "无法获取最新版本"
    fi

    info "最新版本: $version"

    # 构建下载 URL
    local binary_name="${BINARY_NAME}"
    if [ "$os" = "windows" ]; then
        binary_name="${BINARY_NAME}.exe"
    fi

    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}_${platform}"
    if [ "$os" = "windows" ]; then
        download_url="${download_url}.exe"
    fi

    info "下载 ${binary_name}..."
    curl -sL "$download_url" -o "$binary_name"
    chmod +x "$binary_name"

    echo "$binary_name"
}

# 安装二进制文件
install_binary() {
    local binary=$1

    info "安装到 ${INSTALL_DIR}..."
    sudo mv "$binary" "${INSTALL_DIR}/${BINARY_NAME}"
    sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    info "安装完成: ${INSTALL_DIR}/${BINARY_NAME}"
}

# 安装 Skill
install_skill() {
    info "安装 Skill..."

    mkdir -p "$SKILL_DIR"

    # 下载 SKILL.md
    curl -sL "https://raw.githubusercontent.com/${REPO}/main/skill/SKILL.md" -o "${SKILL_DIR}/SKILL.md"

    info "Skill 已安装到: ${SKILL_DIR}"
}

# 验证安装
verify_install() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        info "验证安装..."
        "$BINARY_NAME" version
        info "安装成功！"
    else
        warn "二进制文件已安装，但可能需要重启终端或添加 ${INSTALL_DIR} 到 PATH"
    fi
}

# 主函数
main() {
    echo "=========================================="
    echo "  yh-olap-cli 安装脚本"
    echo "=========================================="
    echo

    local platform
    platform=$(detect_platform)
    info "检测到平台: $platform"

    local binary
    binary=$(download_latest "$platform")

    install_binary "$binary"
    install_skill
    verify_install

    echo
    info "使用方法:"
    echo "  1. 登录: yh-olap-cli login login -u <用户名> -p <密码>"
    echo "  2. 查询: yh-olap-cli query run \"SELECT * FROM table LIMIT 10\""
    echo "  3. 下载: yh-olap-cli download query \"SELECT * FROM table\" -o result.xlsx"
    echo
}

main "$@"
