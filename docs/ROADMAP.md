# logsift 商业化上线开发文档（v1.0 路线）

> 本文档把 logsift 从 "能跑的脚本级工具" 推进到 "敢在生产环境推荐给用户" 的标准。
> 与本文件配套的还有 [DEVELOPMENT_PLAN.md](DEVELOPMENT_PLAN.md)（按难度交替排序的可执行任务列表）。

---

## 0. TL;DR

当前代码（[cmd/logsift/main.go](../cmd/logsift/main.go)、[internal/](../internal/)）实现的是单文件 NDJSON、固定字段、一次扫完即退出的最小可用版本。要达到 "用户开箱就敢在生产用"，必须补齐三件事：

1. **生产场景必备**：follow 模式、多文件/压缩输入、宽松时间戳、稳定的退出码语义。
2. **可信赖**：版本/构建信息、错误诊断、可观测性（自我可调试）、安全边界（路径、内存上限）。
3. **可分发**：跨平台 release artifacts、补全脚本、配置文件 + profile、文档站点。

下文按 **MVP（v1.0 必交付） → v1.x 增值 → v2 远期** 分层，每一项都给出用户故事、CLI 形状、关键设计决策、验收标准。

---

## 1. 产品定位

| 维度 | 表述 |
|---|---|
| 一句话 | **`grep` 太浅，`jq` 太啰嗦的中间地带** —— 面向结构化日志的本地、零依赖、能 piping、能 follow、可交互的过滤器。 |
| 不做什么 | 不做长驻收集（Vector/Fluent Bit 的事），不做存储/索引（Loki/ES 的事），不做告警平台。 |
| 竞品坐标 | 比 `jq` 友好（不写 DSL），比 `lnav` 轻（单二进制无 schema），比 `stern` 通用（不仅 k8s）。 |

## 2. 用户画像与核心场景

| 画像 | 场景 | 当前能否满足 |
|---|---|---|
| **SRE / On-call** | 故障中 `tail -f` 一台机器的 json 日志，按 `trace_id`、`level=error` 快速收敛 | ❌ 无 follow，无嵌套字段 |
| **后端开发** | 本地复现 bug，对着 `app.log` 翻 10 分钟前的请求 | ✅ 基本能用 |
| **数据/平台工程** | 一次性筛一整天的归档日志（gz），导出 TSV 给下游 | ⚠️ 需要先解压、单文件、无 `--until` |
| **支持/客户成功** | 客户发来一坨日志附件，想 5 秒内回答"有没有 timeout" | ✅ 能用，但缺正则、缺统计 |

## 3. Gap 分析（vs "敢上线"标准）

| 类别 | 缺口 | 严重度 |
|---|---|---|
| 输入 | 无 follow、无多文件、无 gzip、单一 NDJSON 格式 | 🔴 阻断 SRE 场景 |
| 解析 | [internal/parser/entry.go:62-66](../internal/parser/entry.go#L62-L66) 注释里就承认 naive timestamp 没支持；非 JSON 行直接丢 | 🔴 数据丢失 |
| 过滤 | `--exclude` 是 [README.md:31](../README.md#L31) 的 TODO；`--where` 不支持嵌套字段、布尔组合、正则 | 🟠 表达力不足 |
| 输出 | 没有 `--fields`、没有 `--count`、退出码不区分"匹配为 0" | 🟠 不适合脚本/CI |
| TUI | [internal/cli/cli.go:95-108](../internal/cli/cli.go#L95-L108) 把全部 kept 装内存才进 TUI；无 detail 视图、无 follow | 🟠 大文件 OOM |
| 工程 | 无 release pipeline、无补全、无配置文件、无版本元信息（git sha/build time） | 🔴 用户拿不到二进制 |
| 可观测 | 无 `--debug`、无 metrics、无 panic recover | 🟡 出问题没法 self-diag |
| 安全 | 路径未校验（symlink/`..`）、无内存/行长上限可配 | 🟡 |
| 文档 | 只有 [README.md](../README.md)，没有 example gallery、cheatsheet、配置参考 | 🟠 |

---

## 4. v1.0 必交付清单（MVP for 公测）

### 4.1 输入层

#### 4.1.1 `--follow` / `-f`（持续跟随）

**用户故事**：on-call 同学打开终端，`logsift --file /var/log/app.json --follow --level=error,warn`，发版后等着新报警往外跳。

**CLI**：
```
--follow, -f                持续读文件尾部，遇 EOF 不退出
--follow-name               文件被 rotate 后按文件名重新打开（默认 inode 跟踪）
--poll-interval=200ms       fsnotify 不可用时的兜底轮询间隔
```

**关键决策**：
- 优先用 `fsnotify`（已是 Go 生态事实标准），Windows 也支持。
- rotate 处理两种语义：默认 inode（跟着旧文件读完再切），`--follow-name` 切到新文件（类似 `tail -F`）。
- TUI 接 follow：底部 stick-to-bottom，按 `f` 切换 follow/freeze。

**改动面**：[internal/cli/cli.go:80-108](../internal/cli/cli.go#L80-L108) 里的 scanner 循环要抽象成 `source.Reader` 接口；新增 `internal/source/{file,follow,stdin}.go`。

**验收**：
- 单测：在 tempdir 模拟 append、rotate-rename、rotate-truncate 三种姿势。
- 手动：`tail -f` 行为对齐，CI 跑一个 60s timeout 的集成脚本验证。

#### 4.1.2 多文件 / glob 输入

**CLI**：
```
logsift app-*.log.gz --level=error
logsift --file shard1.log --file shard2.log --merge-by=ts
```

**关键决策**：
- positional args 也接受文件（更符合 *nix 习惯）；`--file` 保留向后兼容。
- 多文件按时间戳归并（k-way merge），无时间戳的退化为顺序拼接，并在 stderr 打一行提示。

**改动面**：[internal/cli/cli.go:42](../internal/cli/cli.go#L42) 的 `File string` → `Files []string`；新增 `internal/source/merge.go`。

#### 4.1.3 透明解压（gzip / zstd）

按文件后缀（`.gz`、`.zst`）自动 wrap reader；stdin 用 `--decompress=gzip` 显式声明。
关键库：`compress/gzip`（标准库）、`github.com/klauspost/compress/zstd`（Go 纯实现，无 CGO）。

---

### 4.2 解析层

#### 4.2.1 宽松时间戳

[internal/parser/entry.go:63-66](../internal/parser/entry.go#L63-L66) 现在只认 RFC3339。**新增层级**：

```go
var timestampLayouts = []string{
    time.RFC3339Nano,
    time.RFC3339,
    "2006-01-02T15:04:05.000",   // naive ms
    "2006-01-02T15:04:05",        // naive s
    "2006-01-02 15:04:05.000",
    "2006-01-02 15:04:05",
    time.RFC1123Z,
}
```

外加：
- 数值类（unix s / ms / us / ns，按数量级嗅探）。
- naive 时间戳走 `--assume-tz=Local|UTC|+08:00`（默认 `Local`，给个明确的语义，不再静默丢成 zero time）。

**改动面**：[internal/parser/entry.go:33-38](../internal/parser/entry.go#L33-L38) 的 ts 解析；`Options` 增加 `AssumeTZ`。

#### 4.2.2 容错行：非 JSON 不再静默丢弃

[internal/cli/cli.go:85-88](../internal/cli/cli.go#L85-L88) 当前对解析失败直接 `fmt.Fprintf(stderr, "skip: ...")`。生产里"日志里混了一行 stack trace 续行"是常事。

**新行为**：
```
--on-parse-error=skip|warn|keep|fail   默认 warn（≈现状）
```
- `keep`：把整行包成 `Entry{Raw: line, Message: line}`，level 设为 `unknown`，让 grep / output 仍可工作。
- `fail`：第一条不合法就退出 1（CI/管道里有用）。

#### 4.2.3 字段别名 & 自动嗅探

很多框架日志字段叫 `severity` / `lvl` / `@timestamp` / `log.level`（ECS）。

**设计**：内置 `internal/parser/aliases.go`，默认别名表：
```
ts:        ts, time, timestamp, @timestamp, eventTime
level:     level, lvl, severity, log.level
service:   service, app, component, logger
message:   msg, message, log.message
```
可通过 `--field-map=level=severity,service=app` 临时覆盖；或写在配置文件里。

---

### 4.3 过滤层

| 项目 | CLI | 说明 |
|---|---|---|
| `--exclude` 落地 | `--exclude-level=debug,info` `--exclude-grep=heartbeat` | 关掉 [README.md:31](../README.md#L31) 的 TODO |
| 正则 grep | `--grep-re='timeout\|deadline'` | 与 `--grep` 互斥 |
| 跨字段 grep | `--grep-in=msg,error,stack`（默认 `msg`） | |
| 上界时间 | `--until=2h` 或 `--until=2026-05-18T10:00:00` | 与 `--since` 形成区间 |
| 嵌套字段 | `--where 'http.status>=500'`、`tags[0]==prod` | 见 4.3.1 |
| 布尔组合 | `--where 'level==error && http.status>=500'` | 单条表达式内支持 `&&`/`\|\|`/`()`；多条 `--where` 仍为 AND |
| `--invert` / `-v` | 整条匹配链取反 | |

#### 4.3.1 嵌套字段 dot-path

[internal/filter/expr.go:51-64](../internal/filter/expr.go#L51-L64) 的 `lookup` 只看顶层 + 一层 `Extra`。改成递归，支持 `a.b.c` dot-path、数组下标 `tags[0]`。注意 `--where` 解析器要先识别 `[`、`]`、字符串字面量 `"…"`（解决数字误判，见 4.3.2）。

#### 4.3.2 表达式类型显式化

```
version=="1.0"   # 强制字符串比较
status>=500      # 数值（同现状）
flag==true       # bool
```

引号包裹强制走字符串，避免现在 [internal/filter/expr.go:43-48](../internal/filter/expr.go#L43-L48) 的"两边能 ParseFloat 就走数值"导致的歧义。

---

### 4.4 输出层

| 项目 | CLI | 说明 |
|---|---|---|
| 字段投影 | `--fields=ts,level,trace_id,msg` | 对 tsv/json/template 都生效 |
| 模板输出 | `--template='{{.ts}} {{.level}} {{.trace_id}}'` | `text/template`，时间用 `{{.ts.Format "..."}}` |
| 计数模式 | `--count` / `-c` | 只输出匹配总数，配合 `--stats=level` 出分组计数 |
| 退出码语义 | 0=匹配≥1，1=运行错误，2=参数错误，**3=无匹配**（grep 习惯） | 当前 [cmd/logsift/main.go:14-19](../cmd/logsift/main.go#L14-L19) 还没区分 |
| 颜色策略 | `--color=auto\|always\|never`，遵守 `NO_COLOR` env | 现在 [internal/cli/cli.go:65](../internal/cli/cli.go#L65) `useColor` 永远是 `Output=="color"`，没看 TTY |
| 进度/汇总 | `--summary` 末尾打一行 `matched=42 skipped=3 scanned=10000 in 230ms` | 写到 stderr 不污染 stdout |

---

### 4.5 TUI 升级（[internal/tui/model.go](../internal/tui/model.go)）

| 项目 | 说明 |
|---|---|
| 流式装载 | 不再先 buffer 全部 kept 再进 TUI（[internal/cli/cli.go:95-98](../internal/cli/cli.go#L95-L98) 的设计）。改为 `tea.Cmd` 后台读取，ring buffer 默认上限 `--tui-buffer=100000` |
| Detail 视图 | `Enter` 打开右侧 / 下方 pane，渲染完整 JSON（按层级缩进、字段着色） |
| Follow 模式 | TUI 内 `f` 切换，stick-to-bottom；新行红点提示 |
| 高亮 | grep 命中片段 `lipgloss` 反色 |
| 时间跳转 | `g`/`G` 跳首尾，`t` 输入时间戳跳转 |
| 复制到剪贴板 | `y` 复制当前行 raw JSON（用 `golang.design/x/clipboard`） |
| 鼠标 | 滚轮 + 点选 |
| 自适应 level 列表 | [internal/tui/model.go:108](../internal/tui/model.go#L108) 写死的 order 改成根据数据采样 |
| 帮助页 | `?` 弹完整快捷键 modal |

---

### 4.6 工程化

#### 4.6.1 配置文件 & profile

加载顺序：CLI flag > env (`LOGSIFT_*`) > `./.logsift.toml` > `~/.config/logsift/config.toml`。

```toml
default_output = "color"
assume_tz = "Asia/Shanghai"

[fields]
level = "severity"

[profile.prod-errors]
level = "error,fatal"
since = "30m"
exclude_grep = "healthcheck"

[profile.trace]
fields = ["ts", "trace_id", "msg"]
output = "tsv"
```

CLI：`logsift --profile=prod-errors --file …`

库选择：`github.com/BurntSushi/toml`（小、无依赖）。新建 `internal/config/config.go`。

#### 4.6.2 版本/构建信息

[internal/cli/cli.go:20](../internal/cli/cli.go#L20) 现在是硬编码 `const Version = "0.2.0"`。改成 build-time 注入：

```go
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
)
```
Makefile：
```
LDFLAGS = -X github.com/alive-worker/logsift/internal/cli.Version=$(VERSION) \
          -X github.com/alive-worker/logsift/internal/cli.Commit=$(shell git rev-parse --short HEAD) \
          -X github.com/alive-worker/logsift/internal/cli.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
```
`--version` 输出三行：version / commit / build date / go version / os-arch。

#### 4.6.3 Release pipeline（goreleaser）

新增 `.goreleaser.yaml`，CI（`.github/workflows/release.yml`）在 tag 推送时跑：
- 平台：linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64。
- 产物：tar.gz / zip + checksums.txt + SBOM (`syft`) + cosign 签名。
- 自动发 Homebrew tap（`alive-worker/homebrew-tap`）、scoop bucket。
- Docker 镜像推 ghcr.io，多架构 manifest。

#### 4.6.4 Shell 补全

`logsift completion bash|zsh|fish|powershell` 子命令；goreleaser 自动打包到 release tar 里。

#### 4.6.5 自我可观测

```
--debug                  额外日志到 stderr（含 timing、每个 filter 的 in/out 计数）
--trace-file=/tmp/x.json 写更结构化的 trace（用 slog）
```
panic recover 包到 [cmd/logsift/main.go](../cmd/logsift/main.go) 最外层，crash 时打印一份"请把这段贴到 issue 里"的报告（含版本/参数/平台）。

---

### 4.7 非功能性

| 项目 | 目标 | 验证 |
|---|---|---|
| 性能 | 单核 ≥ 500 MB/s NDJSON 扫描（无过滤） | `go test -bench` baseline，CI 上跑回归 |
| 内存 | 流式模式 RSS < 50MB（不论文件大小） | follow + 大文件场景压测 |
| 行长上限 | 默认 1 MB（现状 [internal/cli/cli.go:81](../internal/cli/cli.go#L81)），可调 `--max-line=8M` | |
| 安全 | `--file` 路径校验：拒绝跨 `--root` 的 symlink（仅当显式启用） | 单测覆盖 |
| 隐私 | `--redact='email,phone,/regex/'` 在输出前脱敏（合规友好） | |
| 退出码 | 0/1/2/3 语义稳定（见 4.4） | 集成测试 |

---

## 5. v1.x 增值功能（公测后 3~6 个月）

| 模块 | 功能 | 价值 |
|---|---|---|
| 聚合 | `--stats=field` 分组计数 / `--top=N` / 直方图（控制台 sparkline） | 现场快速洞察 |
| 关联 | `--group-by=trace_id`：把同一 trace 的多条聚成一组 | 微服务排障 |
| 二阶段 | `logsift query save` / `logsift query run` 命名查询 | 团队共享筛选 |
| 远程源 | `--from=loki://…?query=` / `--from=k8s://ns/pod` / `--from=journalctl` | 取代部分 `stern` / `logcli` |
| 写回 | `--export-loki=…` / `--export-clickhouse=…` | 数据回灌 |
| Lua 钩子 | `--plugin=path.lua` 自定义 transform | 给重度用户的逃生舱 |
| LSP 模式 | `logsift lsp` 给编辑器做日志高亮/跳转 | IDE 内体验 |

---

## 6. 远期（v2）

- **嵌入式 SQL 子集**：`logsift sql "SELECT level, count(*) FROM logs WHERE ts > now() - '1h' GROUP BY level"`，用 DuckDB embedding 实现，仅当用户加 `--sql` 子命令时才加载，保持核心二进制小。
- **Web UI**：`logsift serve --port 7575`，把 TUI 的能力暴露成浏览器界面（本地优先，零认证），方便屏幕共享。
- **告警 sidecar**：`logsift watch --rule=rules.yaml --notify=webhook://...`，但要明确"不是 Prometheus"。

---

## 7. 详细设计：跨模块影响地图

```
新增/拆分包：
  internal/source/        file, stdin, follow, gzip, zstd, merge   ← 4.1 全部入口
  internal/parser/        aliases.go, timestamps.go (拆出)          ← 4.2
  internal/filter/expr/   tokenizer.go, parser.go, eval.go          ← 4.3
  internal/output/        template.go, fields.go, summary.go        ← 4.4
  internal/tui/           detail.go, follow.go, help.go             ← 4.5
  internal/config/        config.go, profile.go                     ← 4.6.1
  internal/buildinfo/     version.go                                ← 4.6.2

需要重构的现有文件：
  cmd/logsift/main.go               加 panic recover、统一退出码
  internal/cli/cli.go               Options 字段大幅扩展；Run 拆分 source/sink
  internal/parser/entry.go          timestamp + aliases + on-error 策略
  internal/filter/expr.go           换成 pratt parser 支持 && ||
```

---

## 8. 上线验收 Checklist

- [ ] `logsift --help` ≤ 一屏；高级项放到 `logsift --help-all`
- [ ] README 配 GIF demo（asciinema）；中英文双语保留
- [ ] 文档站（mkdocs-material）部署到 `logsift.dev` 或 GitHub Pages
- [ ] 全部 CLI flag 在文档站有对应章节 + 至少 1 个示例
- [ ] CI：vet + test + race + bench-regression + lint（`golangci-lint`）
- [ ] 主二进制 ≤ 10MB，启动到第一行输出 ≤ 50ms
- [ ] `logsift --version` 输出 commit/build date
- [ ] `panic` 兜底 + crash report
- [ ] `NO_COLOR`、`TERM=dumb` 都被正确尊重
- [ ] 安全：依赖每月跑 `govulncheck`
- [ ] 隐私声明：默认不发任何遥测；如未来引入需 opt-in 且明确写在 README

---

## 9. 风险与权衡

| 风险 | 缓解 |
|---|---|
| 表达式语法越加越像 jq | v1 锁死语法表，给"逃生舱"= `--template` + `--plugin`（v1.x），核心不滑坡 |
| TUI 流式 + follow 复杂度爆炸 | 先做 ring buffer + freeze，按下 `f` 才进 follow；用 bubbletea 的 `tea.Cmd` 划清边界 |
| 多平台 follow 语义差异 | Windows 没 inode → 默认 `--follow-name`；测试矩阵覆盖三个 OS |
| 性能优化诱惑 | 先用标准库 `encoding/json`，等 bench 有 baseline 再换 jsoniter/simdjson；不为优化而优化 |
