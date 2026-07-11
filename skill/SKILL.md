---
name: yh-olap
version: 2.0.0
description: "使用 yh-olap-cli 命令行工具进行 OLAP 数据查询、SQL 执行、结果下载和引擎管理。适用于数据查询、分析报表生成、数据导出、用户登录认证等场景。采用按需登录策略：直接执行命令，仅在认证失败时提示登录。"
metadata:
  requires:
    bins: ["yh-olap-cli"]
  cliHelp: "yh-olap-cli --help"
---

# yh-olap-cli Skill

YH OLAP 命令行工具 Skill - 用于 OLAP 数据查询与下载

## 1. 何时使用本 Skill

### 1.1 触发条件

以下场景应使用本 skill：

- 用户需要执行 SQL 查询 OLAP 数据库
- 用户需要下载查询结果为 Excel 文件
- 用户需要登录/登出 OLAP 系统
- 用户需要查看可用的查询引擎
- 用户提到 yh-olap-cli、OLAP 查询、数据导出、报表生成等关键词

### 1.2 前置条件

1. 首次使用前必须完成用户登录
2. 默认查询引擎为 `impala`，可切换为 `hive` 或 `clickhouse`
3. 支持多个用户凭据，可设置默认用户

## 2. 命令导航

### 2.1 模块地图

| 大模块 | 处理什么问题 | 包含的命令 |
|-------|-------------|-----------|
| 登录管理 | 用户登录、登出、凭据管理 | `login login`、`logout`、`list` |
| SQL 执行 | 提交 SQL 任务（异步） | `exec run`、`exec file` |
| 结果获取 | 获取已执行 SQL 的结果 | `result get` |
| 直接查询 | 执行 SQL 并立即获取结果（同步） | `query run`、`query file` |
| 结果下载 | 下载查询结果为 Excel | `download result`、`download query`、`download file` |
| 引擎管理 | 查看可用查询引擎 | `engines list` |

### 2.2 登录管理模块

用于用户登录认证和凭据管理。

| 命令 | 用途 / 何时使用 | 参数说明 | 示例 |
|------|----------------|----------|------|
| `yh-olap-cli login login` | 登录 OLAP 系统 | `--user/-u` 用户名<br>`--password/-p` 密码<br>`--otp/-o` OTP 密钥<br>`--save/--no-save` 是否保存凭据 [默认: save]<br>`--default/-d` 设为默认用户 | `yh-olap-cli login login -u zhangsan -p xxx -d` |
| `yh-olap-cli logout` | 删除保存的用户凭据 | `[USERNAME]` 要登出的用户名<br>`--all/-a` 清除所有用户 | `yh-olap-cli logout zhangsan`<br>`yh-olap-cli logout --all` |
| `yh-olap-cli list` | 列出所有保存的用户 | - | `yh-olap-cli list` |

### 2.3 SQL 执行模块（异步）

用于提交 SQL 任务，不等待结果返回，适合长时间运行的查询。

| 命令 | 用途 / 何时使用 | 参数说明 | 示例 |
|------|----------------|----------|------|
| `yh-olap-cli exec run` | 执行 SQL 语句（异步） | `SQL` 要执行的 SQL 语句 [必需]<br>`--engine/-e` 查询引擎 [默认: impala]<br>`--user/-u` 用户名<br>`--wait/--no-wait` 等待执行完成 [默认: wait] | `yh-olap-cli exec run "SELECT * FROM table LIMIT 10"` |
| `yh-olap-cli exec file` | 从文件执行 SQL（异步） | `SQL_FILE` SQL 文件路径 [必需]<br>`--engine/-e` 查询引擎 [默认: impala]<br>`--user/-u` 用户名<br>`--wait/--no-wait` 等待执行完成 [默认: wait] | `yh-olap-cli exec file query.sql -e hive` |

> **注意**：异步执行会返回 `request_id`，后续需要使用 `result get` 来获取结果。

### 2.4 结果获取模块

用于获取已提交 SQL 任务的执行结果。

| 命令 | 用途 / 何时使用 | 参数说明 | 示例 |
|------|----------------|----------|------|
| `yh-olap-cli result get` | 获取 SQL 执行结果 | `REQUEST_ID` SQL 执行请求 ID [必需]<br>`--user/-u` 用户名<br>`--format/-f` 输出格式 (table/json/csv) [默认: table]<br>`--limit/-l` 显示行数限制 [默认: 20] | `yh-olap-cli result get req_123 -f json -l 100` |

### 2.5 直接查询模块（同步）

执行 SQL 并立即获取结果，适合快速查询。

| 命令 | 用途 / 何时使用 | 参数说明 | 示例 |
|------|----------------|----------|------|
| `yh-olap-cli query run` | 执行 SQL 并获取结果 | `SQL` 要执行的 SQL 语句 [必需]<br>`--engine/-e` 查询引擎 [默认: impala]<br>`--user/-u` 用户名<br>`--format/-f` 输出格式 [默认: table]<br>`--limit/-l` 显示行数限制 [默认: 20] | `yh-olap-cli query run "SELECT count(*) FROM users" -f csv` |
| `yh-olap-cli query file` | 从文件执行 SQL 并获取结果 | `SQL_FILE` SQL 文件路径 [必需]<br>其他参数同上 | `yh-olap-cli query file report.sql -o result.csv` |

### 2.6 结果下载模块

将查询结果下载为 Excel 文件。

| 命令 | 用途 / 何时使用 | 参数说明 | 示例 |
|------|----------------|----------|------|
| `yh-olap-cli download result` | 下载已有 SQL 执行结果为 Excel | `REQUEST_ID` SQL 执行请求 ID [必需]<br>`--output/-o` 输出文件路径<br>`--user/-u` 用户名<br>`--engine/-e` 查询引擎 | `yh-olap-cli download result req_123 -o report.xlsx` |
| `yh-olap-cli download query` | 执行 SQL 并下载结果 | `SQL` 要执行的 SQL 语句 [必需]<br>`--output/-o` 输出文件路径<br>`--engine/-e` 查询引擎<br>`--user/-u` 用户名 | `yh-olap-cli download query "SELECT * FROM sales" -o sales.xlsx` |
| `yh-olap-cli download file` | 从文件执行 SQL 并下载结果 | `SQL_FILE` SQL 文件路径 [必需]<br>`--output/-o` 输出文件路径<br>其他参数同上 | `yh-olap-cli download file monthly.sql -o monthly.xlsx` |

### 2.7 引擎管理模块

查看可用的查询引擎。

| 命令 | 用途 / 何时使用 | 参数说明 | 示例 |
|------|----------------|----------|------|
| `yh-olap-cli engines list` | 列出所有可用的查询引擎 | - | `yh-olap-cli engines list` |

## 3. 通用知识与最佳实践

### 3.1 查询引擎选择

| 引擎 | 特点 | 适用场景 |
|-----|------|---------|
| `impala` | 默认引擎，速度快 | 交互式查询、快速数据分析 |
| `hive` | 稳定性好，支持复杂查询 | 大规模数据处理、ETL 任务 |
| `clickhouse` | 列式存储，聚合查询性能优异 | 大数据量统计分析 |

### 3.2 典型工作流

**快速查询（推荐）**：
```bash
# 一步执行并获取结果
yh-olap-cli query run "SELECT * FROM table LIMIT 100"
```

**大查询 + 下载**：
```bash
# 方式1：直接查询并下载（适合中等大小结果）
yh-olap-cli download query "SELECT * FROM big_table" -o result.xlsx

# 方式2：异步执行 + 下载（适合超大查询）
yh-olap-cli exec run "SELECT * FROM very_big_table" --no-wait
# 记录返回的 request_id，稍后获取结果
yh-olap-cli download result req_xxx -o result.xlsx
```

**多用户切换**：
```bash
# 登录并保存为默认用户
yh-olap-cli login login -u user1 -p xxx -d

# 使用非默认用户
yh-olap-cli query run "SELECT * FROM table" -u user2
```

### 3.3 输出格式说明

| 格式 | 特点 | 适用场景 |
|-----|------|---------|
| `table` | 美观的表格显示 | 终端查看、小结果集 |
| `json` | 结构化数据 | 程序处理、数据交换 |
| `csv` | 逗号分隔值 | Excel 导入、进一步处理 |
| Excel | 通过 `download` 命令 | 报表生成、业务用户使用 |

## 4. 执行规则

### 4.1 标准执行顺序

1. **直接执行命令**：不需要预先检查登录状态，直接执行用户请求的操作
2. 根据任务类型选择合适的命令模块
   - 快速查看结果 → `query run` / `query file`
   - 需要 Excel 文件 → `download query` / `download file`
   - 长时间运行的大查询 → `exec run` + `download result`
3. 选择合适的查询引擎（默认 impala，特殊场景可切换）
4. 设置合适的输出格式和限制
5. 执行命令并处理结果
6. **仅在认证失败时**：提示用户需要登录并引导完成登录流程

### 4.2 不可违反规则

1. **不预先检查登录**：不要在执行命令前主动检查登录状态，直接执行命令
2. **失败后再引导登录**：只有当命令返回认证/权限相关错误时，才提示用户登录
3. **大查询不要用默认 limit**：当用户需要全量数据时，明确指定 `--limit` 或使用 `download` 命令
4. **默认用户优先**：未指定用户时使用默认用户
5. **异步执行需记录 request_id**：使用 `exec` 命令后，保存返回的 request_id 供后续查询
6. **注意 SQL 引号转义**：复杂 SQL 建议使用 `file` 方式执行

### 4.3 输出限制

- `result get` 和 `query` 系列命令默认只显示前 20 行
- 需要更多行时显式指定 `--limit N`
- 超过 1000 行建议直接下载为 Excel 文件

## 5. 常见错误与恢复

| 错误 / 现象 | 含义 | 恢复动作 |
|------------|------|----------|
| **认证失败 / 权限错误 / 未登录** | 用户未登录或凭据过期 | **这是最常见的错误**。当命令执行返回认证相关错误时，主动提示用户需要登录，并提供登录命令示例：<br>`yh-olap-cli login login -u 用户名 -p 密码 [-d]`<br>登录成功后重新执行原命令即可。 |
| 引擎不支持 | 指定的引擎不存在 | 执行 `yh-olap-cli engines list` 查看可用引擎 |
| SQL 语法错误 | SQL 语句有问题 | 检查 SQL 语法，复杂 SQL 建议写入文件用 `file` 方式执行 |
| 查询超时 | 查询执行时间过长 | 使用 `exec` 异步模式，或优化 SQL 语句 |
| request_id 不存在 | 任务 ID 错误或已过期 | 重新提交查询，或检查 request_id 是否正确 |
| 文件不存在 | SQL 文件路径错误 | 检查文件路径，使用绝对路径或正确的相对路径 |

### 5.1 认证失败处理流程

当执行数据操作命令遇到认证/登录相关错误时，严格按照以下流程处理：

1. **识别认证错误**：从命令输出中识别出是登录/认证/权限相关问题
2. **向用户说明**：告知用户当前需要登录才能继续操作
3. **提供登录命令**：给出清晰的登录命令示例，包含参数说明
4. **登录成功后重试**：用户完成登录后，自动或提示重新执行原来的命令

**示例处理话术**：
> 检测到需要登录才能访问 OLAP 系统。请使用以下命令登录：
> ```bash
> yh-olap-cli login login -u 你的用户名 -p 你的密码 -d
> ```
> 参数说明：`-d` 表示设为默认用户，后续操作无需重复指定用户名。
> 登录成功后即可重新执行查询。
