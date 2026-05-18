.PHONY: build test vet fmt run docker docker-shell submission clean

BINARY         := logsift
PKG            := ./cmd/logsift
SUBMISSION_DIR := submissions/logsift

build:
	go build -o $(BINARY) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

run: build
	./$(BINARY) --file testdata/sample.ndjson --output tsv

docker:
	docker build -t logsift .

docker-shell:
	docker run --rm -it logsift bash -c 'cd /app && bash'

# Regenerate the submission package from current HEAD.
# Dockerfile is mirrored from the repo root (single source of truth).
# repo.zip is produced via `git archive` so its contents always track HEAD,
# excluding paths marked `export-ignore` in .gitattributes.
# `git -c core.autocrlf=false -c core.eol=lf` forces LF endings on Windows
# checkouts; would otherwise rewrite text files to CRLF in the archive.
submission:
	@mkdir -p $(SUBMISSION_DIR)
	@git diff-index --quiet HEAD -- || \
		echo "warn: working tree is dirty; repo.zip reflects HEAD, not unstaged changes" >&2
	cp Dockerfile $(SUBMISSION_DIR)/Dockerfile
	git -c core.autocrlf=false -c core.eol=lf archive \
		--format=zip --output=$(SUBMISSION_DIR)/repo.zip HEAD
	@echo "regenerated: $(SUBMISSION_DIR)/Dockerfile + $(SUBMISSION_DIR)/repo.zip"

clean:
	rm -f $(BINARY) $(BINARY).exe
	go clean -testcache
