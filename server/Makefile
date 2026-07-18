.PHONY: build test lint stop clean remove

COMPOSE_FILE ?= ../docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

build:
	$(COMPOSE) run --rm server go build ./...

test:
	$(COMPOSE) run --rm server go test ./...

lint:
	$(COMPOSE) run --rm server sh -c 'test -z "$$(gofmt -l .)" || (echo "Go files are not formatted:"; gofmt -l .; exit 1)'
	$(COMPOSE) run --rm server go vet ./...

stop:
	$(COMPOSE) stop

clean:
	$(COMPOSE) down --remove-orphans

remove: clean
