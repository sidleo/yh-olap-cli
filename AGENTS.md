# AGENTS.md

## Build & Run

```bash
go build -o yh-olap-cli .
./yh-olap-cli --help
```

Cross-compile: `GOOS=windows GOARCH=amd64 go build -o yh-olap-cli.exe .`

No Makefile, no CI, no tests yet.

## Architecture

Cobra CLI (`cmd/`) + internal packages (`internal/`). Single binary, no runtime deps.

- `main.go` → `cmd.Execute()` → cobra root
- `cmd/*.go` — one file per command group (login, exec, query, result, download, engines)
- `internal/auth/` — CAS login + TOTP + config (~/.yh_olap/config.json)
- `internal/api/` — HTTP client for OLAP backend (sql, download, approval)
- `internal/engine/` — Hive/Impala/Clickhouse definitions
- `internal/output/` — table/json/csv formatters
- `skill/SKILL.md` — agent skill definition

## CAS Login Quirks (internal/auth/yhlogin.go)

This is the trickiest part of the codebase:

- **Cookie jar is mandatory** — two `http.Client` instances share one jar; without it the session state breaks and login fails silently.
- **Two clients**: one follows redirects (to discover login URL), one does NOT (to capture 302 from POST).
- **Form data fields `phoneNum`, `captcha`, `geolocation` must be OMITTED** — Go's `url.Values` sends empty strings for `""`, but the CAS server treats empty strings differently from absent fields (Python httpx sends `None` = absent). Sending them causes 401.
- **Ticket redirect**: OTP POST returns `Location: https://oa?ticket=ST-xxx`. Extract the ticket and append it to the **original** `OLAPServiceURL`, not the `oa` URL.
- **SSL**: `InsecureSkipVerify: true` is required (internal cert chain issue).

## Config Compatibility

`~/.yh_olap/config.json` format matches the Python `yh_olap` package. Passwords are base64-encoded. Both versions can coexist and share credentials.

## Skill Installation

The install script outputs the skill directory path. Agents must copy the skill to their own skills directory. The script does NOT auto-detect or auto-install to agent-specific paths.

## What's Missing vs Python yh_olap

- Interactive REPL (`yh-olap-cli i`) — directory `internal/repl/` exists but is empty
- HDFS operations (moveToTrash, getHdfsDir, uploadSingleFile)
- Cluster management (killJob)
- CSV-to-Hive SQL generation
- Tests (no `*_test.go` files exist)
