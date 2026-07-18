.PHONY: build test lint

build:
	go build ./...

test:
	go test ./...

lint:
	@test -z "$$(gofmt -l .)" || (echo "Go files are not formatted:"; gofmt -l .; exit 1)
	go vet ./...
