# yh-olap-cli

永辉 OLAP 数据查询命令行工具。用于 SQL 执行、结果查询、Excel 下载等场景。

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/sidleo/yh-olap-cli/main/install.sh | bash
```

安装完成后，将 Skill 复制到您的 Agent skills 目录：

```bash
# 查看 Skill 文件位置（安装脚本会输出）
ls ~/.agents/skills/yh-olap-cli/SKILL.md

# 复制到您的 Agent skills 目录（以 Claude Code 为例）
mkdir -p ~/.claude/skills/yh-olap-cli
cp ~/.agents/skills/yh-olap-cli/SKILL.md ~/.claude/skills/yh-olap-cli/
```

## 使用

### 登录

```bash
yh-olap-cli login login -u <用户名> -p <密码> -o <OTP密钥> -d
```

| 参数 | 说明 |
|------|------|
| `-u, --user` | 用户名 |
| `-p, --password` | 密码 |
| `-o, --otp` | OTP 密钥 |
| `-d, --default` | 设为默认用户 |

### 查询

```bash
# 执行 SQL 并获取结果
yh-olap-cli query run "SELECT * FROM table LIMIT 10"

# 指定输出格式
yh-olap-cli query run "SELECT * FROM table" -f json
yh-olap-cli query run "SELECT * FROM table" -f csv

# 从文件执行
yh-olap-cli query file query.sql
```

### 下载

```bash
# 执行 SQL 并下载为 Excel
yh-olap-cli download query "SELECT * FROM table" -o result.xlsx

# 下载已有结果
yh-olap-cli download result <requestId> -o result.xlsx
```

### 其他命令

```bash
# 查看可用引擎
yh-olap-cli engines list

# 列出已保存用户
yh-olap-cli list

# 登出
yh-olap-cli logout <用户名>
```

## 查询引擎

| 引擎 | 说明 |
|------|------|
| `impala` | 默认引擎，速度快 |
| `hive` | 稳定性好，支持复杂查询 |
| `clickhouse` | 列式存储，聚合查询性能优异 |

```bash
# 指定引擎
yh-olap-cli query run "SELECT * FROM table" -e hive
```

## Agent 集成

安装后，Agent 可直接使用 `yh-olap-cli` 命令。建议的使用流程：

1. **按需登录**：直接执行查询命令，认证失败时再提示登录
2. **大查询下载**：超过 1000 行建议用 `download` 命令
3. **复杂 SQL**：使用 `file` 方式避免引号转义问题

## 配置

配置文件位于 `~/.yh_olap/config.json`，与 Python 版本 `yh_olap` 兼容。

## 开发

```bash
# 构建
go build -o yh-olap-cli .

# 交叉编译
GOOS=windows GOARCH=amd64 go build -o yh-olap-cli.exe .
```

## License

MIT
