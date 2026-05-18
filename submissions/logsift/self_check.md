# §3.3 验证环境 — self-check (logsift)

All seven items verified on Docker 29.4.3, image `logsift:trial`.

| # | Check                                | Status | Evidence                                                                          |
|---|--------------------------------------|--------|-----------------------------------------------------------------------------------|
| 1 | 镜像可成功构建                        | PASS   | `docker build -t logsift:trial .` → `naming to docker.io/library/logsift:trial done`. |
| 2 | 容器可正常启动                        | PASS   | `docker run --rm logsift:trial bash -c 'echo ok'` → `ok`.                         |
| 3 | 当前工作目录为 `/app`                 | PASS   | `pwd` → `/app`.                                                                   |
| 4 | `repo` 存在                          | PASS   | `ls -A /app` → `.git Dockerfile README.md cmd go.mod internal testdata`. No host residue. |
| 5 | `repo` 为 Git 仓库                   | PASS   | `git rev-parse --is-inside-work-tree` → `true`.                                   |
| 6 | 容器内代码为任务起始现场              | PASS   | `git log --oneline` → single `initial scene` commit; `git status --short` empty. |
| 7 | 进入容器后无需额外手工初始化即可开始工作 | PASS | `go version` → go1.24.x; `go build ./...` clean; `go test ./...` → cli/filter/parser/tui packages all OK. |

## Reproduce locally

```bash
docker build -t logsift:trial .
docker run --rm logsift:trial bash -c '
  pwd
  git -C /app log --oneline
  git -C /app status --short
  go test ./...
'
```

Expected last line: three `ok` lines plus two `[no test files]` lines.

## Operator note

Use `bash -c` rather than `bash -lc` when invoking the shell.
Debian's `/etc/profile` (run by `-l`) resets `PATH` and strips
`/usr/local/go/bin`, which would make `go` look missing despite being
installed at the standard location. `bash -c` keeps the image's ENV PATH.

## Build mirrors

Dockerfile uses `goproxy.cn` for Go modules and Tsinghua TUNA for apt
to make builds reliable in mainland-China networks; both are upstream
mirrors with identical content.
