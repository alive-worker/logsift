# logsift v1.0 详细开发计划

> 配套文档：[ROADMAP.md](ROADMAP.md)。本文件把路线图拆成 **24 个可独立 PR 的任务**，按 **简单(S) → 中等(M) → 困难(H) 严格交替** 排序，目的是让团队节奏稳定：
>
> - 每完成一个困难任务，下一轮先做一个简单任务回血。
> - 困难任务始终有简单/中等任务"挂载"在它前后，便于在卡壳时切换上下文继续推进。
> - 简单任务往往直接面向用户可见的改进，能持续产生 release note 素材；困难任务则在背后铺地基。

---

## 1. 难度分级标准

| 等级 | 标记 | 衡量口径 | 单人工时 |
|---|---|---|---|
| 简单 | **S** | 单文件 / 单包改动；逻辑直接；不需新依赖 | ≤ 1 人天 |
| 中等 | **M** | 跨包改动 / 需新增包 / 引入 1 个新依赖；需要写新测试套件 | 2~5 人天 |
| 困难 | **H** | 涉及架构调整 / 跨平台行为差异 / 外部基建（CI、签名、发布）；需 design review | ≥ 1 人周 |

判定原则：**风险点的数量** 比 **代码量** 更主导难度。一个 50 行的 fsnotify 跨平台代码可能是 H，一个 200 行的 lookup 函数仍是 S。

---

## 2. 总览：24 任务交替路线

| # | 难度 | 任务 | 输出 | 里程碑 | 主要依赖 |
|---|---|---|---|---|---|
| 01 | **S** | 版本/构建信息 LDFLAGS 注入 | `--version` 显示 commit / build date | M1 | 无 |
| 02 | **M** | source/sink 抽象重构 | `internal/source`、`internal/sink` 接口落地 | M1 | 01 |
| 03 | **H** | Goreleaser 发布管线（多平台 + Homebrew + Docker manifest） | tag → 多架构二进制 + 镜像 | M1 | 01 |
| 04 | **S** | 退出码语义 + `--count` | 0/1/2/3 标准化 | M1 | 02 |
| 05 | **M** | 多文件 / glob 输入 + 按 ts k-way 归并 | positional args 接文件 | M1 | 02 |
| 06 | **H** | TUI 流式 + ring buffer | 大文件不再 OOM | M1 | 02 |
| 07 | **S** | `--exclude-level` / `--exclude-grep` | 落实 README TODO | M2 | 02 |
| 08 | **M** | gzip / zstd 透明解压 | `app.log.gz` 直接读 | M2 | 02 |
| 09 | **H** | `--follow` / `-f` + rotate 跨平台 | tail -f 替代 | M2 | 02, 06 |
| 10 | **S** | 宽松时间戳 + `--assume-tz` + 字段别名表 | naive ts 不再被丢 | M2 | 无 |
| 11 | **M** | `--where` 嵌套 dot-path + 数组下标 | `http.status>=500` 可用 | M2 | 无 |
| 12 | **H** | 表达式 Pratt parser：`&&` `\|\|` `()` + 引号字符串 | 单条 `--where` 内布尔组合 | M2 | 11 |
| 13 | **S** | `--until` 上界时间 + `--invert`/`-v` + `--max-count` | 时间区间 / 反选 / 截断 | M2 | 10 |
| 14 | **M** | `--fields` 投影 + `--template`（text/template） | 自定义输出 | M2 | 02 |
| 15 | **H** | 配置文件 + profile + 加载优先级 | `~/.config/logsift/config.toml` | M3 | 02 |
| 16 | **S** | `--color=auto/always/never` + NO_COLOR + panic recover | 颜色策略合规、crash 有报告 | M3 | 无 |
| 17 | **M** | 正则 grep `--grep-re` + 跨字段 `--grep-in` | 表达力补齐 | M3 | 02 |
| 18 | **H** | 自观测：`--debug` / `--trace-file` / `--summary` + govulncheck CI | slog trace、漏洞扫描 | M3 | 02 |
| 19 | **S** | TUI 帮助 modal (`?`) + grep 高亮 + 动态 level 列表 | TUI 体验细节 | M3 | 06 |
| 20 | **M** | TUI detail 视图 + 时间跳转 + 复制行 | `Enter` 打开 JSON 详情 | M3 | 06 |
| 21 | **H** | TUI 接 follow + stick-to-bottom + 新行提示 | follow 与 TUI 闭环 | M3 | 09, 06 |
| 22 | **S** | Shell 补全（bash/zsh/fish/powershell） | `logsift completion …` | M3 | 03 |
| 23 | **M** | 聚合：`--stats=field` + `--top=N` + sparkline | 分组计数 | M3 | 02 |
| 24 | **H** | 文档站 mkdocs-material + cosign 签名 + SBOM | 发布 GA | M3 | 03 |

> 节奏校验：把 #1~#24 的难度列出来是 `S M H S M H S M H S M H S M H S M H S M H S M H` —— **严格的 S/M/H 三拍循环**。

---

## 3. 任务详情

> 每项都包含：用户故事 / CLI 形状 / 主要改动文件 / 验收 / 依赖 / 估时。
> 改动文件路径相对仓库根目录。

### Task 01 [S] — 版本/构建信息 LDFLAGS 注入
- **用户故事**：用户把 issue 贴上来时附 `logsift --version`，我们能立刻知道哪个 commit。
- **CLI**：`logsift --version` 输出 4 行：`logsift v1.0.0` / `commit: abc1234` / `built: 2026-05-18T…` / `go1.26 darwin/arm64`。
- **改动**：
  - [internal/cli/cli.go](../internal/cli/cli.go)：`const Version = "0.2.0"` → 三个 `var`（Version/Commit/BuildDate）。
  - 新建 [internal/buildinfo/info.go](../internal/buildinfo/info.go) 汇总，包含 `runtime.Version()`、`GOOS`、`GOARCH`。
  - [Makefile](../Makefile)：`LDFLAGS = -X 'github.com/alive-worker/logsift/internal/buildinfo.Version=$(VERSION)' …`。
- **验收**：
  - `make build VERSION=v1.0.0-test` 后 `./logsift --version` 输出包含传入字符串。
  - dev 构建（直接 `go build`）应得到默认 `dev` / `unknown`，不报错。
- **依赖**：无。**估时**：0.5 天。

---

### Task 02 [M] — source / sink 抽象重构
- **用户故事**：没有直接用户故事，是后续 follow / 多文件 / TUI 流式的地基。
- **设计**：
  ```go
  // internal/source/source.go
  type Source interface {
      Next(ctx context.Context) (line []byte, err error)  // io.EOF 表示读完；可阻塞（follow 模式）
      Close() error
  }
  // internal/sink/sink.go — 把 output.Writer 包装一层，便于 count/summary 注入计数器
  ```
- **改动**：
  - 拆分 [internal/cli/cli.go:80-110](../internal/cli/cli.go#L80-L110) 的 scanner 循环：
    - 输入侧 → `internal/source/{stdin,file}.go`。
    - 输出侧 → `internal/sink/sink.go`（包装 `output.Writer` + 计数）。
    - 主循环改为 `for line := range source.Next() { … }`，统一关闭与错误传播。
  - `cli.Run` 接收 `context.Context`，最外层 `cmd/logsift/main.go` 处理 SIGINT。
- **验收**：
  - 现有所有单元测试保持绿。
  - 新增 `internal/source/stdin_test.go` 验证 EOF/取消语义。
  - benchmark：扫描 100k 行 stdin，吞吐相对重构前回归 ≤ 5%。
- **依赖**：01。**估时**：4 天。

---

### Task 03 [H] — Goreleaser 发布管线
- **用户故事**：`brew install alive-worker/tap/logsift` 或 `scoop install logsift`，3 秒装好；docker `pull ghcr.io/alive-worker/logsift` 直接能用。
- **改动**：
  - 新建 [.goreleaser.yaml](../.goreleaser.yaml)：
    - `builds`：5 个目标（linux amd64/arm64、darwin amd64/arm64、windows amd64）。
    - `archives`：tar.gz + zip，含 `README.md`、`LICENSE`、自动生成的 `completions/`。
    - `nfpms`：apt/yum 包（linux）。
    - `brews`：推 `alive-worker/homebrew-tap`。
    - `dockers` + `docker_manifests`：multi-arch image 推 ghcr.io。
    - `signs`：cosign keyless（OIDC）；`sboms`：syft 生成 SPDX。
  - 新建 [.github/workflows/release.yml](../.github/workflows/release.yml)：tag `v*` 触发；权限 `contents: write` `packages: write` `id-token: write`。
  - GitHub secrets：`HOMEBREW_TAP_TOKEN`（PAT）。
- **验收**：
  - 推一个 `v0.3.0-rc.1` tag，CI 跑通，5 个平台二进制 + Homebrew + Docker manifest + cosign 签名都到位。
  - `cosign verify-blob --bundle …` 通过。
  - `brew install --HEAD` 能跑通 `--version`。
- **依赖**：01（要有版本变量）。**估时**：7 天（首次配置 + 3 次试错周期）。

---

### Task 04 [S] — 退出码语义 + `--count`
- **用户故事**：CI/脚本里 `if logsift --grep panic --since 1h; then alert; fi` 能稳定工作。
- **CLI**：
  ```
  --count, -c       仅输出匹配总数到 stdout
  退出码：0=匹配≥1；1=运行错误；2=参数错误；3=无匹配
  ```
- **改动**：
  - [cmd/logsift/main.go](../cmd/logsift/main.go)：区分 `cli.ErrUsage`（→2）、`cli.ErrNoMatch`（→3）、其他 error（→1）。
  - `Options.Count bool`；`sink.Sink` 维护 matched 计数。
  - 文档：把退出码列入 README。
- **验收**：
  - 集成测试 4 种退出码各 1 例（含 `--count` + 无匹配）。
- **依赖**：02。**估时**：1 天。

---

### Task 05 [M] — 多文件 / glob 输入 + k-way 归并
- **用户故事**：`logsift app-2026-05-*.log --since=24h` 一次性筛一周归档。
- **CLI**：positional args 接文件；`--file` 改可重复并保留。
- **改动**：
  - [internal/cli/cli.go:42](../internal/cli/cli.go#L42) `File string` → `Files []string`；非 flag args 也 append。
  - 新建 `internal/source/multi.go`：k-way merge by `Entry.Timestamp`（min-heap）。
  - 无时间戳条目降级为"按 Files 顺序拼接"，stderr 提示一行。
  - Glob 在 OS 不展开时由代码 `filepath.Glob` 兜底（Windows cmd 必备）。
- **验收**：
  - 单测覆盖：3 文件交错 ts；其中 1 文件无 ts；glob 命中 0 文件 → 错误退出 2。
- **依赖**：02。**估时**：3 天。

---

### Task 06 [H] — TUI 流式 + ring buffer
- **用户故事**：把 5GB 文件拖给 `logsift --tui`，不会 OOM；最近 10 万行随时可看。
- **改动**：
  - [internal/cli/cli.go:95-108](../internal/cli/cli.go#L95-L108) 当前 `kept []*Entry` 全量缓存的设计废弃。
  - TUI 内部维护 `RingBuffer[*Entry]`，容量 `--tui-buffer=100000`（默认）。
  - 后台 goroutine 通过 `tea.Cmd` 把 `source.Next()` 的结果以 batch 形式推送（`tea.Msg`），主线程仅做渲染。
  - 背压：当 buffer 满 + 用户停止滚动时丢弃最旧；底部 status 栏显示 `dropped=N`。
- **验收**：
  - 1GB 文件压测 RSS < 200MB（基线现状 ≈ 文件大小）。
  - 测试：[internal/tui/model_test.go](../internal/tui/model_test.go) 新增 stream / drop 用例。
- **依赖**：02。**估时**：8 天。

---

### Task 07 [S] — `--exclude-level` / `--exclude-grep`
- **用户故事**：排查时不想看 debug 噪声：`--exclude-level=debug`；不想看心跳：`--exclude-grep=heartbeat`。
- **改动**：
  - [internal/filter/filter.go](../internal/filter/filter.go) 增加 `ExcludeLevelFilter` / `ExcludeGrepFilter`（或给现有 Filter 加 `negate bool`，更优雅）。
  - [internal/cli/cli.go:139-162](../internal/cli/cli.go#L139-L162) `buildChain` 增加两个 flag 处理。
  - 删除 [README.md:31](../README.md#L31) 的 "reserved for follow-up" 一行。
- **验收**：单测 + README 示例。
- **依赖**：02。**估时**：1 天。

---

### Task 08 [M] — gzip / zstd 透明解压
- **用户故事**：`logsift app.log.gz` 直接读；归档场景常态。
- **改动**：
  - `internal/source/decompress.go`：按后缀返回包装后的 `io.Reader`；显式 `--decompress=gzip|zstd|none|auto`。
  - go.mod：`+github.com/klauspost/compress/zstd`。
  - 多文件混合（部分 gz）：在 `multi.go` 里逐个判断。
- **验收**：
  - 单测：3 文件（1 plain + 1 gz + 1 zst）交错 ts。
  - `--decompress=none` 对 `.gz` 文件应得到乱码（验证不强制嗅探时的行为）。
- **依赖**：02。**估时**：3 天。

---

### Task 09 [H] — `--follow` / `-f` + rotate 跨平台
- **用户故事**：`logsift app.log -f --level=error,warn` 替代 `tail -f app.log | grep error`，且日志切割不掉行。
- **设计**：
  - `internal/source/follow.go`：
    - Linux/macOS：fsnotify + inode 跟踪。rotate-rename（如 logrotate）下读完旧 inode → reopen 新 inode。
    - Windows：fsnotify 不报 rename 给当前句柄；默认 `--follow-name` 语义（按文件名重打开）。
    - tmpfs / NFS / fsnotify 失败：fallback 到 `--poll-interval=200ms` 轮询。
  - 不支持的源（stdin、gzip）显式拒绝 `-f`，退出码 2。
- **CLI**：见 ROADMAP §4.1.1。
- **验收**：
  - 集成测试矩阵（CI 上跑 Linux + Windows + macOS）：append / rotate-rename / rotate-truncate / 长时间空闲（>30s）。
  - 60s timeout 的端到端脚本：写 100 行 → rotate → 再写 100 行 → 应读到全部 200 行。
- **依赖**：02、06（TUI 接 follow 在 Task 21）。**估时**：10 天（跨平台调试占大头）。

---

### Task 10 [S] — 宽松时间戳 + `--assume-tz` + 字段别名表
- **用户故事**：`@timestamp` / `severity` / naive ts 一上来就能解析，不用调参。
- **改动**：
  - [internal/parser/entry.go:63-66](../internal/parser/entry.go#L63-L66)：扩 layouts 列表 + unix 数值嗅探。
  - 新建 `internal/parser/aliases.go`：默认别名映射 + `--field-map` 覆盖。
  - `Options.AssumeTZ string`（默认 `Local`）。
- **验收**：
  - [internal/parser/entry_test.go](../internal/parser/entry_test.go) 增加 10+ 种 ts 格式覆盖。
  - `ECS` 字段（`@timestamp` / `log.level`）自动识别用例。
- **依赖**：无。**估时**：1 天。

---

### Task 11 [M] — `--where` 嵌套 dot-path + 数组下标
- **用户故事**：`logsift --where 'http.status>=500' --where 'tags[0]==prod'`。
- **改动**：
  - [internal/filter/expr.go:51-64](../internal/filter/expr.go#L51-L64) `lookup` → 递归 + 数组下标解析。
  - 新建 `internal/filter/expr/lookup.go`：单元化测试。
- **验收**：
  - 单测覆盖：3 层嵌套；数组越界返回 nil；中间节点不是 map/slice 返回 nil。
- **依赖**：无（独立于 Task 12 的 parser 重写）。**估时**：3 天。

---

### Task 12 [H] — `--where` 表达式 Pratt parser
- **用户故事**：`--where 'level==error && (http.status>=500 || latency>1000)'`。
- **设计**：
  - `internal/filter/expr/tokenizer.go`：识别标识符、数字、字符串（`"…"`）、`==/!=/<=/>=/</>`、`&&`、`||`、`!`、`(`、`)`。
  - `internal/filter/expr/parser.go`：Pratt parser 输出 AST。
  - `internal/filter/expr/eval.go`：AST 求值器，复用 Task 11 的 `lookup`。
  - 兼容：旧 `field<op>value` 语法仍能解析（单 token 节点）。
- **CLI**：单条 `--where` 内支持布尔；多条 `--where` 仍 AND。
- **验收**：
  - 单测 + property-based fuzz（`testing/fuzz`）：随机表达式不能 panic。
  - 错误信息友好：定位列号，比如 `--where: expected operand at column 17`。
- **依赖**：11。**估时**：8 天。

---

### Task 13 [S] — `--until` + `--invert`/`-v` + `--max-count`
- **用户故事**：`--since=2h --until=1h` 取一小时窗口；`-v --grep heartbeat` 排除；`--max-count=100` 截断。
- **改动**：
  - `internal/filter/filter.go`：增加 `UntilFilter`、链层包 `Negate(Filter)`。
  - `cli.Run` 在写入时 short-circuit `--max-count`。
- **验收**：单测 + 与 `--count` 组合（应在到达 max 后立刻停 source）。
- **依赖**：10（`--until` 解析复用 since 的时间格式）。**估时**：1 天。

---

### Task 14 [M] — `--fields` + `--template`
- **用户故事**：`logsift --fields=ts,level,trace_id --output=tsv`；`--template='{{.ts}} [{{.trace_id}}] {{.msg}}'`。
- **改动**：
  - `internal/output/fields.go`：解析 `--fields`，对 tsv/json 都生效（json 输出 reduced object）。
  - `internal/output/template.go`：`text/template`，预注入函数 `now`、`color`、`pad`、`truncate`。
  - 命令行：`--output=template` 触发模板路径。
- **验收**：
  - 模板编译错在 `cli.ParseArgs` 阶段就报错（不是运行到第一行才报）。
  - 单测覆盖各种字段缺失（应渲染为 `<nil>` 而不是 panic）。
- **依赖**：02。**估时**：3 天。

---

### Task 15 [H] — 配置文件 + profile + 加载优先级
- **用户故事**：团队约定 `~/.config/logsift/config.toml` 存常用 profile，跨成员一致。
- **改动**：
  - 新建 `internal/config/{config.go,profile.go}`：TOML schema + 合并逻辑。
  - 优先级：CLI flag > `LOGSIFT_*` env > `./.logsift.toml` > `$XDG_CONFIG_HOME/logsift/config.toml`。
  - `--profile=name`：在合并后再叠加 profile section（profile 内的字段覆盖外层默认）。
  - `--print-config`：打印最终生效配置（调试用）。
- **风险**：合并语义复杂，容易引入"我以为它读了，但它没读"的体验问题 → 必须有 `--print-config` 作为逃生舱。
- **验收**：
  - 加载矩阵：12 种组合（4 来源 × 3 字段类型）单元测试。
  - 文档：明确写出加载顺序 + `--print-config` 示例。
- **依赖**：02。**估时**：7 天。

---

### Task 16 [S] — `--color=auto/always/never` + NO_COLOR + panic recover
- **用户故事**：管道时不出现 ANSI 乱码；crash 时有可粘贴的报告。
- **改动**：
  - [internal/cli/cli.go:65](../internal/cli/cli.go#L65) `useColor` 改成根据 `--color` + `isatty(stdout)` + `NO_COLOR` env 决定。
  - [cmd/logsift/main.go](../cmd/logsift/main.go) 最外层 `defer recover()`，打印 version + args + goroutine stack 到 stderr，退出码 134。
- **验收**：
  - `NO_COLOR=1 logsift … > /dev/null` 后 stdout 无 ESC。
  - 单测：注入一个故意 panic 的 filter，验证 main 不裸 crash。
- **依赖**：无。**估时**：1 天。

---

### Task 17 [M] — 正则 grep + 跨字段
- **用户故事**：`--grep-re='timeout|deadline'`、`--grep-in=msg,error,stack`。
- **改动**：
  - `internal/filter/filter.go`：`RegexGrepFilter`；`Options.GrepIn []string`。
  - `--grep` / `--grep-re` 互斥（ParseArgs 校验）。
- **验收**：单测 + benchmark（正则 vs 子串性能基线）。
- **依赖**：02。**估时**：2 天。

---

### Task 18 [H] — 自观测：`--debug` / `--trace-file` / `--summary` + govulncheck
- **用户故事**：用户报 "为什么我的 `--where` 没生效"，让他加 `--debug` 一眼看到 chain 上每个 filter 的 in/out 计数。
- **改动**：
  - 全局换用 `log/slog`，trace 走 JSON handler。
  - `internal/sink/sink.go` 维护 per-filter counter（通过 chain 装饰器），结束时 `--summary` 输出到 stderr。
  - `.github/workflows/ci.yml` 增加 `govulncheck` job。
- **验收**：
  - `--debug --summary` 输出：每个 filter `level / since / where[0] / …` 的 keep/drop 计数。
  - govulncheck CI 在已知漏洞依赖时失败。
- **依赖**：02。**估时**：6 天。

---

### Task 19 [S] — TUI 帮助 modal + grep 高亮 + 动态 level 列表
- **用户故事**：第一次开 TUI 的用户按 `?` 看到所有快捷键；搜 `timeout` 时命中片段被反色。
- **改动**：
  - [internal/tui/model.go](../internal/tui/model.go) 新增 overlay 状态机；`?` 切换。
  - [internal/tui/model.go:107-116](../internal/tui/model.go#L107-L116) `cycleLevel` 改为基于实际数据采样的 level 列表。
  - 渲染时把 grep 命中片段用 `lipgloss.NewStyle().Reverse(true)` 包裹。
- **验收**：[internal/tui/model_test.go](../internal/tui/model_test.go) 加 3 个用例。
- **依赖**：06。**估时**：1 天。

---

### Task 20 [M] — TUI detail 视图 + 时间跳转 + 复制行
- **用户故事**：选中一行按 `Enter` 看到完整 JSON 缩进展示；按 `t` 输入时间戳跳到附近行；按 `y` 复制原 JSON。
- **改动**：
  - 新建 `internal/tui/detail.go`：JSON pretty print（按字段着色）。
  - 输入框组件（`bubbles/textinput`）做时间戳输入。
  - 剪贴板：`golang.design/x/clipboard`，初始化失败时 graceful disable + 提示。
- **验收**：单测 + 手动 demo gif 录入文档。
- **依赖**：06。**估时**：4 天。

---

### Task 21 [H] — TUI 接 follow + stick-to-bottom + 新行提示
- **用户故事**：TUI 中按 `f` 开启 follow，新 error 来时屏幕底部红点+计数闪烁。
- **改动**：
  - 把 Task 09 的 `source.Follow` 接入 Task 06 的 ring buffer，统一通过 `tea.Cmd` 流入。
  - 状态：`follow` / `frozen`；frozen 时来的新行计入"未读"计数。
  - 退出语义：`q` 退出时，把当前 visible 切片转写到 stdout（如果原本被管道捕获），不丢数据。
- **验收**：
  - 集成测试：spawn `logsift -f ... --tui`，外部 append，验证 UI 状态机变化（用 bubbletea 的 testable 模式）。
- **依赖**：06、09。**估时**：8 天。

---

### Task 22 [S] — Shell 补全
- **用户故事**：`logsift --le<TAB>` 自动补成 `--level=`。
- **改动**：
  - 若仍用 `flag` 包，手写补全脚本（`scripts/completions/{bash,zsh,fish,ps1}`）；如果决定换 cobra，则自带。
  - goreleaser archives 段引用补全脚本。
- **验收**：
  - macOS + Linux + Windows 各装一遍，回车前后补全行为符合预期。
- **依赖**：03。**估时**：1 天。

---

### Task 23 [M] — 聚合：`--stats=field` + `--top=N` + sparkline
- **用户故事**：`logsift app.log --stats=service --top=10` 5 秒看出请求量最大的服务。
- **改动**：
  - 新建 `internal/sink/stats.go`：sink 模式之一；hash map 计数；结束时按数量排序输出。
  - Sparkline：用 Unicode block characters `▁▂▃▄▅▆▇█` 按时间分桶渲染。
  - 与 `--count` 互斥。
- **验收**：
  - 单测 + 性能：100 万条 stats 应 < 1s（内存 hashmap）。
- **依赖**：02。**估时**：4 天。

---

### Task 24 [H] — 文档站 + cosign 签名验证 + SBOM
- **用户故事**：用户访问 `logsift.dev`（GitHub Pages）查所有 flag；安全敏感场景的用户能 `cosign verify` 二进制。
- **改动**：
  - `docs/site/`：mkdocs-material 配置 + Cheatsheet + 全 flag 参考。
  - `.github/workflows/docs.yml`：push main 后部署 Pages。
  - 把 Task 03 留下的 cosign keyless 签名补到文档站 "Verify the release" 页。
  - SBOM 验证脚本：`scripts/verify-sbom.sh`。
- **验收**：
  - 站点上线，所有页面通过 `lychee` 死链检查。
  - 第三方按文档步骤 `cosign verify-blob --bundle …` 成功。
- **依赖**：03。**估时**：6 天。

---

## 4. 里程碑映射

| 里程碑 | 时间窗 | 包含任务 | 用户可感知的标志 |
|---|---|---|---|
| **M1 — v0.3 "Tail"** | 第 1~3 周 | 01, 02, 03, 04, 05, 06 | 用户能拿到正式签名的二进制 / 多文件 / TUI 不 OOM / 退出码可脚本化 |
| **M2 — v0.4 "Expression"** | 第 4~7 周 | 07, 08, 09, 10, 11, 12, 13, 14 | `tail -f` 取代场景成立 / `--where` 表达力对齐 jq 子集 |
| **M3 — v1.0 "Polish"** | 第 8~12 周 | 15~24 | 配置文件 + 自观测 + TUI follow + 文档站 + 补全，全部完成 |

> 看板建议：每个 milestone 在 GitHub Projects 设一列；每个任务一张 issue，标题前缀 `[S/M/H]`，便于按难度筛选。

---

## 5. 并行性与依赖图

```
01 ── 02 ─┬─ 03 ─── 22 ── 24
          ├─ 04
          ├─ 05
          ├─ 06 ─┬── 19
          │      ├── 20
          │      └── 21
          ├─ 07
          ├─ 08
          ├─ 09 ──┘
          ├─ 14
          ├─ 15
          ├─ 17
          ├─ 18
          └─ 23

10 ── 13
11 ── 12
16 (无依赖)
```

**可并行起跑**：在 02 完成后，{04, 05, 07, 08, 14, 15, 17, 18, 23} 全部独立可并行；{10, 11, 13, 16} 任何时候都能起。

---

## 6. 验收门禁（每个 PR 必过）

- [ ] `go vet ./... && go test ./...` 全绿（CI 已配 [.github/workflows/ci.yml](../.github/workflows/ci.yml)）。
- [ ] 新增公共 API 有单元测试覆盖 ≥ 80%。
- [ ] 影响用户行为的改动有对应 [README.md](../README.md) 章节更新。
- [ ] 影响性能路径的改动附 benchmark 对比数（贴在 PR 描述）。
- [ ] 跨平台敏感改动（follow、路径、补全）有 Linux + macOS + Windows 三平台 CI runner 验证。
- [ ] 困难任务（H）必须有 design doc，PR 描述中链接到 `docs/designs/<task-id>.md`。

---

## 7. 风险登记表（动态维护）

| ID | 风险 | 触发任务 | 缓解 | 状态 |
|---|---|---|---|---|
| R1 | Windows fsnotify 不报 rename | 09 | 默认 `--follow-name` + 显式文档 | 待开始 |
| R2 | 表达式语法滑坡向 jq | 12 | v1.0 锁语法表；新功能优先走 `--template` | 待开始 |
| R3 | goreleaser cosign keyless 跑不通 | 03 | 先 fallback 到 `gpg` 离线签名；后续切 keyless | 待开始 |
| R4 | TUI 流式 + follow 状态机 bug | 06+21 | 写 testable 模式集成测试；先发 RC 收社区反馈 | 待开始 |
| R5 | 配置文件优先级体感歧义 | 15 | 提供 `--print-config` 调试出口 | 待开始 |

---

## 8. 估时汇总

| 类别 | 任务数 | 总人天 |
|---|---|---|
| S（简单） | 8 | 6.5 |
| M（中等） | 8 | 26 |
| H（困难） | 8 | 60 |
| **合计** | **24** | **92.5 人天** |

按单人节奏约 **18 周**；2 人并行（H 任务不并行同一条依赖链）约 **10~11 周**，与 M1+M2+M3 的 12 周窗口对齐。
