# logsift

> [English](#english) ｜ [中文](#中文)

---

## English

`logsift` filters newline-delimited JSON logs from stdin or a file. It is a
small local tool for the common day-to-day case where `grep` is too shallow
(matches across fields you didn't intend) and `jq` is too verbose (rewriting
a filter for every quick question).

### Quick start

```bash
go build -o logsift ./cmd/logsift
cat app.log | ./logsift --level=error,warn --since=10m --grep="timeout"
./logsift --file app.log --where 'service==api' --where 'status>=500' --output tsv
```

### Filters

- `--level=<csv>` — accept only entries whose `level` field is in the list.
- `--since=<dur>` — keep entries whose `ts` is within the past duration
  (e.g. `15m`, `2h`, `90s`, `3d`).
- `--grep=<substr>` — substring match against `msg` (case-insensitive).
- `--where=<expr>` — repeatable; each expression is `field<op>value` with
  ops `==`, `!=`, `>=`, `<=`, `>`, `<`. Numeric compare kicks in when both
  sides parse as numbers; otherwise it falls back to string compare.
- `--exclude` is reserved for a follow-up task.

### Output

- `--output=color` (default): coloured single-line summary.
- `--output=json`: passthrough of matched JSON lines.
- `--output=tsv`: tab-separated `ts<TAB>level<TAB>service<TAB>msg`.

### Tests

```bash
go test ./...
```

### Running in Docker

```bash
docker build -t logsift .
docker run --rm -it logsift bash -c '
  cd /app
  go test ./...
'
```

The container's working directory is `/app`, Go toolchain is pre-installed,
and the repo is a single-commit clean initial scene.

> Tip: use `bash -c` rather than `bash -lc`. Debian's `/etc/profile` strips
> `/usr/local/go/bin` from PATH under a login shell.

### Layout

```
cmd/logsift/main.go          entrypoint
internal/cli/                flag parsing, wiring
internal/parser/             NDJSON line parser + duration parser
internal/filter/             filter chain + expression evaluator
internal/output/             color / json / tsv writers
testdata/                    sample logs used by tests
```

---

## 中文

`logsift` 是一个从标准输入或文件读取 NDJSON 日志、并按级别 / 时间窗 / 关键词 /
字段表达式过滤的本地小工具。它面向的日常场景是：`grep` 太浅（会跨字段误匹配），
而 `jq` 又太啰嗦（每个临时问题都得重写过滤式）。

### 快速上手

```bash
go build -o logsift ./cmd/logsift
cat app.log | ./logsift --level=error,warn --since=10m --grep="timeout"
./logsift --file app.log --where 'service==api' --where 'status>=500' --output tsv
```

### 过滤器

- `--level=<csv>` — 只保留 `level` 字段在列表中的条目。
- `--since=<dur>` — 只保留 `ts` 在过去给定时长内的条目（例如 `15m`、`2h`、`90s`、`3d`）。
- `--grep=<substr>` — 对 `msg` 做大小写不敏感的子串匹配。
- `--where=<expr>` — 可重复；每条表达式形如 `字段<op>值`，支持 `==`、`!=`、
  `>=`、`<=`、`>`、`<`。两侧都能解析为数字时走数值比较，否则走字符串比较。
- `--exclude` 预留给后续任务实现。

### 输出

- `--output=color`（默认）：彩色的单行摘要。
- `--output=json`：把匹配到的原始 JSON 行原样输出。
- `--output=tsv`：制表符分隔 `ts<TAB>level<TAB>service<TAB>msg`。

### 跑测试

```bash
go test ./...
```

### 在 Docker 中运行

```bash
docker build -t logsift .
docker run --rm -it logsift bash -c '
  cd /app
  go test ./...
'
```

容器内工作目录是 `/app`，Go 工具链已装好，仓库是一份单 commit 的干净起始现场。

> 注意：调用容器内 shell 时请用 `bash -c`，不要用 `bash -lc`。Debian 的
> `/etc/profile` 在 login shell 下会把 `/usr/local/go/bin` 从 PATH 里去掉，
> 导致 `go` 看起来"不存在"。

### 目录结构

```
cmd/logsift/main.go          入口
internal/cli/                flag 解析 + 装配
internal/parser/             NDJSON 行解析 + duration 解析
internal/filter/             过滤链 + 表达式求值器
internal/output/             color / json / tsv 三种输出
testdata/                    测试用样例日志
```
