# yh-olap_cli 实现方案

## 1. 语言选择：Go

### 推荐理由

| 考量 | Go | Rust |
|------|-----|------|
| 编译为单二进制 | ✅ | ✅ |
| 交叉编译 | ✅ 原生支持，`GOOS/GOARCH` 即可 | ⚠️ 需要 cross 工具链 |
| HTTP 客户端生态 | `net/http` + `encoding/json` 原生够用 | `reqwest` 等外部 crate |
| JSON 处理 | 原生 `encoding/json` + `struct` tag | `serde` 编译时间长 |
| REPL 交互 | `promptui` / `liner` | `rustyline` |
| Excel 生成 | `excelize` (成熟库) | `calamine` (读)/`xlsxwriter` (写) |
| TOTP 生成 | `pquerna/otp` | `totp` crate |
| 编译速度 | 快 | 慢 |
| 学习曲线 | 低 | 高 |
| Agent 环境兼容 | 完美（无运行时依赖） | 完美 |

**结论**：Go 在 HTTP API 客户端场景下生态成熟、编译快、交叉编译简单，是最佳选择。

---

## 2. 认证方案

### 2.1 现有 yhlogin 认证流程分析

```
用户凭证 (username, password, otp_key)
         │
         ▼
┌─────────────────────────────────────┐
│ 1. GET  idaas-cas.yonghui.cn/cas/login?service=... │
│    → 获取登录页面 URL (target_login_url)            │
├─────────────────────────────────────┤
│ 2. GET  target_login_url            │
│    → 初始化 session                     │
├─────────────────────────────────────┤
│ 3. POST target_login_url            │
│    Body: flag=1, username, password,            │
│          execution=e2s1, _eventId=submit        │
│    → 200 (需要OTP) 或 302 (登录成功)           │
├─────────────────────────────────────┤
│ 4. (如果401) POST target_login_url  │
│    Body: flag=1, username,                    │
│          token=pyotp.TOTP(otp_key).now(),      │
│          execution=e2s2, _eventId=submit        │
│    → 302 (登录成功)                           │
├─────────────────────────────────────┤
│ 5. GET  redirect URL (follow 302)   │
│    → 提取 cookies 中的 JSESSIONID              │
└─────────────────────────────────────┘
         │
         ▼
    Token: "JSESSIONID=<jsessionid>"
```

### 2.2 Go 实现方案

需要实现的依赖（仅 3 个外部库）：
- `pquerna/otp` — TOTP 生成（替代 pyotp）
- `excelize` — Excel 读写（替代 pandas + openpyxl）
- `net/http` + `golang.org/x/net/html` — CAS 登录表单解析

SSL 特殊处理：
```go
// yhlogin 使用禁用证书验证的 SSL context
// 原因：YHBI 内部证书链问题
transport := &http.Transport{
    TLSClientConfig: &tls.Config{
        InsecureSkipVerify: true,
    },
}
```

---

## 3. 项目结构

```
yh-olap_cli/
├── main.go                          # 入口
├── go.mod
├── go.sum
├── Makefile                         # 构建 + 交叉编译
├── .goreleaser.yml                  # 发布配置
│
├── cmd/                             # CLI 命令定义
│   └── root.go                      # cobra root 命令
│
├── internal/
│   ├── auth/                        # 认证模块
│   │   ├── yhlogin.go               # CAS 登录流程
│   │   ├── totp.go                  # TOTP 生成
│   │   └── config.go                # 配置读写 (~/.yh_olap/config.json)
│   │
│   ├── api/                         # API 客户端
│   │   ├── client.go                # HTTP 客户端 (token, headers, base_url)
│   │   ├── sql.go                   # SQL 执行 + 结果查询
│   │   ├── download.go              # 下载管理
│   │   ├── approval.go              # 审批订单
│   │   ├── hdfs.go                  # HDFS 操作
│   │   └── cluster.go               # 集群管理
│   │
│   ├── engine/                      # 引擎定义
│   │   └── engine.go                # Hive/Impala/Clickhouse
│   │
│   ├── output/                      # 输出格式化
│   │   ├── table.go                 # 表格输出
│   │   ├── json.go                  # JSON 输出
│   │   ├── csv.go                   # CSV 输出
│   │   └── progress.go              # 进度条
│   │
│   └── repl/                        # 交互模式
│       └── repl.go
│
├── skill/                           # Agent Skill
│   └── SKILL.md
│
├── install.sh                       # 一键安装脚本
│
└── tests/
    ├── auth_test.go
    ├── api_test.go
    └── cli_test.go
```

### 依赖清单

```go
// go.mod
module github.com/yonghui-bigdata/yh-olap_cli

go 1.22

require (
    github.com/spf13/cobra v1.8.0        // CLI 框架
    github.com/pquerna/otp v1.4.0         // TOTP
    github.com/xuri/excelize/v2 v2.8.0   // Excel 读写
    github.com/cheggaaa/pb/v3 v3.2.0     // 进度条
    github.com/AlecAivazis/survey/v2     // 交互式输入
)
```

---

## 4. 实现阶段

### Phase 1: 基础框架 + 认证（3-4天）

**目标**：搭建项目骨架，实现登录/登出/配置管理

```
yh-olap login login -u xxx -p xxx -o xxx
yh-olap login logout [user]
yh-olap login list
```

**任务清单**：
1. 初始化 Go module，添加 cobra 依赖
2. 实现 `internal/auth/config.go` — 配置文件读写（兼容现有 `~/.yh_olap/config.json` 格式）
3. 实现 `internal/auth/yhlogin.go` — CAS 登录流程
   - GET 登录页面 → POST 密码认证 → (可选) POST OTP → 跟随 302 → 提取 JSESSIONID
4. 实现 `internal/auth/totp.go` — TOTP 生成
5. 实现 login/logout/list 命令
6. 编写认证模块单元测试

**验收标准**：
- `yh-olap login login -u <user> -p <pass> -o <otp>` 成功登录
- `yh-olap login list` 显示已保存用户
- `yh-olap login logout <user>` 删除用户
- 配置文件格式与 Python 版本兼容（可互用）

### Phase 2: SQL 执行 + 结果获取（3-4天）

**目标**：实现 SQL 执行、结果查询、同步查询

```
yh-olap exec run "SQL" [-e engine]
yh-olap exec file <path.sql>
yh-olap result get <requestId> [-f table/json/csv]
yh-olap query run "SQL" [-f table/json/csv] [-l limit]
yh-olap query file <path.sql>
```

**任务清单**：
1. 实现 `internal/api/client.go` — HTTP 客户端封装
   - 公共 headers (token, orgCode, Content-Type, User-Agent)
   - 请求/响应处理
   - 错误处理
2. 实现 `internal/api/sql.go` — runSql, getLogResult, getSqlResult, checkState, queryHisSqlResult
3. 实现 `internal/engine/engine.go` — 引擎定义和映射
4. 实现 exec run/file 命令（异步执行，轮询等待）
5. 实现 result get 命令（获取结果并格式化输出）
6. 实现 query run/file 命令（执行+获取结果组合）
7. 实现 `internal/output/` — table/json/csv 格式化
8. 编写集成测试（mock HTTP server）

**验收标准**：
- 执行 SQL 返回 request_id
- 结果以 table/json/csv 格式正确显示
- 支持三种引擎切换
- `--limit` 参数正确限制行数

### Phase 3: 下载功能（2-3天）

**目标**：实现 Excel 下载（含审批流程）

```
yh-olap download result <requestId> -o file
yh-olap download query "SQL" -o file
yh-olap download file <path.sql> -o file
```

**任务清单**：
1. 实现 `internal/api/download.go` — olapResultSimple, olapResult, downloadToExcel, refresh
2. 实现 `internal/api/approval.go` — createMiddleDownloadOrder, createSkipDownloadOrder, detail
3. 实现下载逻辑：
   - <=1000 行：直接 GET `/download/olapResultSimple/{requestId}`
   - \>1000 行：创建审批 → 轮询状态 → 根据引擎选择下载端点
4. 实现 download result/query/file 命令
5. 添加进度条显示

**验收标准**：
- 小结果集直接下载
- 大结果集自动走审批流程
- 进度条正确显示
- 下载的 Excel 文件可用 Microsoft Excel/WPS 打开

### Phase 4: 交互模式 + 引擎管理（1-2天）

**目标**：实现 REPL 和引擎列表

```
yh-olap i
yh-olap engines list
```

**任务清单**：
1. 实现 `internal/repl/repl.go` — 交互式终端
   - SQL 语法提示
   - 历史记录
   - 内置命令（help, exit, engine, engines, result, download）
2. 实现 engines list 命令
3. 添加 `--version` 支持

**验收标准**：
- REPL 中可直接输入 SQL 执行
- 引擎切换正常工作
- 历史记录持久化

### Phase 5: Skill + 安装脚本 + 发布（2-3天）

**目标**：创建 Skill 文件，实现一键安装，配置 GitHub 发布

**任务清单**：
1. 编写 `skill/SKILL.md` — 与现有 yh-olap skill 格式一致
2. 编写 `install.sh` — 检测平台，下载二进制，安装 skill
3. 配置 `.goreleaser.yml` — 多平台交叉编译
4. 配置 GitHub Actions CI/CD
5. 添加 HDFS 和 Cluster API 支持（可选，按需）

**安装流程**：
```bash
# 方式 1: 一键安装
curl -fsSL https://raw.githubusercontent.com/yonghui-bigdata/yh-olap_cli/main/install.sh | bash

# 方式 2: go install
go install github.com/yonghui-bigdata/yh-olap_cli@latest

# 方式 3: 手动下载
# 从 GitHub Releases 下载对应平台二进制
```

---

## 5. 分发策略

### 5.1 主要分发渠道

| 渠道 | 优先级 | 说明 |
|------|--------|------|
| GitHub Releases | P0 | 主要分发方式，goreleaser 自动生成多平台二进制 |
| install.sh 脚本 | P0 | Agent 环境一键安装，自动下载 + 安装 skill |
| go install | P1 | 开发者手动安装 |
| Homebrew tap | P2 | macOS 用户（可选） |

### 5.2 二进制命名规则

```
yh-olap_darwin_arm64     # macOS Apple Silicon
yh-olap_darwin_amd64     # macOS Intel
yh-olap_linux_amd64      # Linux x64
yh-olap_linux_arm64      # Linux ARM
yh-olap_windows_amd64.exe # Windows
```

### 5.3 install.sh 设计

```bash
#!/bin/bash
# 1. 检测 OS + ARCH
# 2. 从 GitHub Releases 下载对应二进制
# 3. 安装到 /usr/local/bin 或 ~/.local/bin
# 4. 安装 Skill 到 ~/.agents/skills/yh-olap/SKILL.md
# 5. 验证安装
```

---

## 6. Skill 设计

Skill 文件与现有 `yh-olap` skill 保持一致，只替换二进制名称：

```yaml
---
name: yh-olap
description: 使用 yh-olap 命令行工具进行 OLAP 数据查询...
metadata:
  requires:
    bins:
      - yh-olap    # 新二进制名称
---
```

Skill 内容复用现有 SKILL.md，更新命令前缀。

---

## 7. 测试策略

| 层级 | 方法 | 覆盖范围 |
|------|------|----------|
| 单元测试 | `go test` | 配置读写、TOTP 生成、输出格式化 |
| 集成测试 | httptest mock server | API 调用、认证流程、下载逻辑 |
| 端到端测试 | 手动 / CI | 完整 CLI 命令执行 |

关键测试场景：
- 认证流程（密码登录 → OTP 验证 → Token 提取）
- 配置文件兼容性（与 Python 版本互用）
- 大结果集下载（审批流程）
- 多引擎 SQL 执行
- REPL 交互

---

## 8. 工作量估算

| 阶段 | 工作量 | 天数 |
|------|--------|------|
| Phase 1: 基础框架 + 认证 | 核心 | 3-4 天 |
| Phase 2: SQL 执行 + 结果 | 核心 | 3-4 天 |
| Phase 3: 下载功能 | 核心 | 2-3 天 |
| Phase 4: REPL + 引擎管理 | 补充 | 1-2 天 |
| Phase 5: Skill + 发布 | 补充 | 2-3 天 |
| **总计** | | **11-16 天** |

如果优先实现核心功能（Phase 1-3），可在 **8-11 天**内完成可用版本。

---

## 9. 兼容性保证

1. **配置文件兼容**：`~/.yh_olap/config.json` 格式完全一致，Python 和 Go 版本可互用
2. **命令行参数兼容**：所有命令和参数与 Python 版本保持一致
3. **API 兼容**：调用相同的后端 API，行为一致
4. **输出格式兼容**：table/json/csv/Excel 输出格式一致

---

## 10. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| CAS 登录流程变更 | 认证失败 | 流程有详细分析，变更概率低；可快速适配 |
| SSL 证书问题 | 连接失败 | 已知需要 `InsecureSkipVerify`，与 Python 版本一致 |
| Excel 格式差异 | 下载文件打不开 | 使用 excelize 库，与 pandas/openpyxl 输出兼容 |
| Agent 沙箱网络限制 | 无法访问 API | 与 Python 版本相同的网络要求 |
