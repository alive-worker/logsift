.PHONY: build test vet fmt run docker docker-shell clean

BINARY := logsift
PKG    := ./cmd/logsift

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

clean:
	rm -f $(BINARY) $(BINARY).exe
	go clean -testcache
