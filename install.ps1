# yh-olap-cli 安装脚本 (Windows)

$ErrorActionPreference = "Stop"

$REPO = "sidleo/yh-olap-cli"
$BINARY_NAME = "yh-olap-cli"
$INSTALL_DIR = "$env:USERPROFILE\.local\bin"
$SKILL_DIR = "$env:USERPROFILE\.agents\skills\yh-olap"

# 颜色输出函数
function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Green }
function Write-Warn { Write-Host "[WARN] $args" -ForegroundColor Yellow }
function Write-Error { Write-Host "[ERROR] $args" -ForegroundColor Red; exit 1 }

# 检测架构
function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Error "不支持的架构: $arch" }
    }
}

# 获取最新版本
function Get-LatestVersion {
    Write-Info "获取最新版本..."
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest" -UseBasicParsing
        return $release.tag_name
    } catch {
        Write-Error "无法获取最新版本: $_"
    }
}

# 下载二进制文件
function Download-Binary {
    param([string]$Version, [string]$Arch)

    $binaryName = "$BINARY_NAME`_windows_$Arch.exe"
    $downloadUrl = "https://github.com/$REPO/releases/download/$Version/$binaryName"
    $outputFile = "$BINARY_NAME.exe"

    Write-Info "下载 $binaryName..."
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $outputFile -UseBasicParsing
        return $outputFile
    } catch {
        Write-Error "下载失败: $_"
    }
}

# 安装二进制文件
function Install-Binary {
    param([string]$BinaryPath)

    # 创建安装目录
    if (-not (Test-Path $INSTALL_DIR)) {
        New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
    }

    Write-Info "安装到 $INSTALL_DIR..."
    Copy-Item $BinaryPath "$INSTALL_DIR\$BINARY_NAME.exe" -Force

    # 添加到 PATH（如果不在其中）
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$INSTALL_DIR*") {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$INSTALL_DIR", "User")
        Write-Warn "已将 $INSTALL_DIR 添加到 PATH，请重启终端生效"
    }

    Write-Info "安装完成: $INSTALL_DIR\$BINARY_NAME.exe"
}

# 安装 Skill
function Install-Skill {
    Write-Info "安装 Skill..."

    if (-not (Test-Path $SKILL_DIR)) {
        New-Item -ItemType Directory -Path $SKILL_DIR -Force | Out-Null
    }

    $skillUrl = "https://raw.githubusercontent.com/$REPO/main/skill/SKILL.md"
    try {
        Invoke-WebRequest -Uri $skillUrl -OutFile "$SKILL_DIR\SKILL.md" -UseBasicParsing
        Write-Info "Skill 已安装到: $SKILL_DIR"
    } catch {
        Write-Warn "Skill 安装失败: $_"
    }
}

# 验证安装
function Verify-Install {
    $binaryPath = "$INSTALL_DIR\$BINARY_NAME.exe"
    if (Test-Path $binaryPath) {
        Write-Info "验证安装..."
        & $binaryPath version
        Write-Info "安装成功！"
    } else {
        Write-Warn "二进制文件已安装，但可能需要重启终端"
    }
}

# 主函数
function Main {
    Write-Host "=========================================="
    Write-Host "  yh-olap-cli 安装脚本 (Windows)"
    Write-Host "=========================================="
    Write-Host ""

    $arch = Get-Arch
    Write-Info "检测到架构: $arch"

    $version = Get-LatestVersion
    Write-Info "最新版本: $version"

    $binary = Download-Binary -Version $version -Arch $arch
    Install-Binary -BinaryPath $binary
    Install-Skill
    Verify-Install

    Write-Host ""
    Write-Info "使用方法:"
    Write-Host "  1. 登录: yh-olap-cli login login -u <用户名> -p <密码>"
    Write-Host "  2. 查询: yh-olap-cli query run `"SELECT * FROM table LIMIT 10`""
    Write-Host "  3. 下载: yh-olap-cli download query `"SELECT * FROM table`" -o result.xlsx"
    Write-Host ""
}

Main
