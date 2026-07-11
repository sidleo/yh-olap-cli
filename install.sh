#!/bin/bash
set -e

REPO="sidleo/yh-olap-cli"
BINARY_NAME="yh-olap-cli"
SKILL_DIR="$HOME/.agents/skills/yh-olap"
INSTALL_DIR="$HOME/.local/bin"

info() { echo -e "\033[0;32m[INFO]\033[0m $1" >&2; }
warn() { echo -e "\033[1;33m[WARN]\033[0m $1" >&2; }
error() { echo -e "\033[0;31m[ERROR]\033[0m $1" >&2; exit 1; }

# 检测平台
detect_platform() {
    local os arch
    case "$(uname -s)" in
        Linux*)  os="linux";;
        Darwin*) os="darwin";;
        *)       error "不支持的操作系统: $(uname -s)"
    esac
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64";;
        arm64|aarch64)  arch="arm64";;
        *)              error "不支持的架构: $(uname -m)"
    esac
    echo "${os}_${arch}"
}

# 获取二进制文件（仅输出路径到 stdout）
get_binary() {
    local platform=$1

    # 尝试从 GitHub Releases 下载
    info "尝试从 GitHub Releases 下载..."
    local version
    version=$(curl -sL --max-time 10 "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -n "$version" ] && ! echo "$version" | grep -q "Not Found"; then
        info "最新版本: $version"
        local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}_${platform}"
        curl -sL --max-time 30 "$download_url" -o "/tmp/${BINARY_NAME}" 2>/dev/null
        if [ -s "/tmp/${BINARY_NAME}" ]; then
            chmod +x "/tmp/${BINARY_NAME}"
            echo "/tmp/${BINARY_NAME}"
            return 0
        fi
    fi

    # 下载失败，从源码构建
    warn "GitHub Releases 不可用，从源码构建..."
    if ! command -v go &> /dev/null; then
        error "未安装 Go，请先安装: brew install go"
    fi

    local build_dir="/tmp/yh-build-$$"
    rm -rf "$build_dir"

    if ! git clone --depth 1 "https://github.com/${REPO}.git" "$build_dir" 2>/dev/null; then
        error "克隆仓库失败"
    fi

    cd "$build_dir"
    if go build -o "$BINARY_NAME" . 2>&1; then
        echo "$build_dir/$BINARY_NAME"
        return 0
    else
        error "构建失败"
    fi
}

# 安装
install_binary() {
    local binary=$1

    mkdir -p "$INSTALL_DIR"
    mv "$binary" "${INSTALL_DIR}/${BINARY_NAME}"
    chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    info "已安装到 ${INSTALL_DIR}/${BINARY_NAME}"
}

# 安装 Skill
install_skill() {
    mkdir -p "$SKILL_DIR"
    curl -sL "https://raw.githubusercontent.com/${REPO}/main/skill/SKILL.md" -o "${SKILL_DIR}/SKILL.md" 2>/dev/null || true
    info "Skill 已安装到 ${SKILL_DIR}"
}

# 主流程
echo "=========================================="
echo "  yh-olap-cli 安装脚本"
echo "=========================================="

platform=$(detect_platform)
info "平台: $platform"

binary=$(get_binary "$platform")
install_binary "$binary"
install_skill

echo
info "验证安装..."
"${INSTALL_DIR}/${BINARY_NAME}" version 2>/dev/null || true

echo
info "使用方法:"
echo "  1. 登录: yh-olap-cli login login -u <用户名> -p <密码>"
echo "  2. 查询: yh-olap-cli query run \"SELECT * FROM table LIMIT 10\""
echo "  3. 下载: yh-olap-cli download query \"SELECT * FROM table\" -o result.xlsx"
