# logsift

`logsift` filters newline-delimited JSON logs from stdin or a file. It is a
small local tool meant for the common day-to-day case where `grep` is too
shallow (matches across fields you didn't intend) and `jq` is too verbose
(rewriting a filter for every quick question).

## Quick start

```bash
go build -o logsift ./cmd/logsift
cat app.log | ./logsift --level=error,warn --since=10m --grep="timeout"
./logsift --file app.log --where 'service==api' --where 'status>=500' --output tsv
```

## Filters

- `--level=<csv>` — accept only entries whose `level` field is in the list.
- `--since=<dur>` — keep entries whose `ts` is within the past duration
  (e.g. `15m`, `2h`, `90s`).
- `--grep=<substr>` — substring match across `msg`.
- `--where=<expr>` — repeatable; each expression is `field<op>value`
  with ops `==`, `!=`, `>=`, `<=`, `>`, `<`. String and numeric compares
  are inferred from the literal.
- `--exclude` is reserved for a follow-up task.

## Output

- `--output=color` (default for TTY): coloured single-line summary.
- `--output=json`: passthrough of matched JSON lines.
- `--output=tsv`: tab-separated `ts<TAB>level<TAB>service<TAB>msg`.

## Layout

```
cmd/logsift/main.go          entrypoint
internal/cli/                flag parsing, wiring
internal/parser/             NDJSON line parser + `since` duration parser
internal/filter/             filter chain + expression evaluator
internal/output/             color / json / tsv writers
testdata/                    sample logs used by tests
```

## Tests

```bash
go test ./...
```
