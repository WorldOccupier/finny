.PHONY: build test lint run stop clean remove

COMPOSE_FILE ?= ../docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)
export DOCKER_BUILDKIT := 1
export COMPOSE_DOCKER_CLI_BUILD := 1

build:
	$(COMPOSE) run --rm server go build ./...

test:
	$(COMPOSE) run --rm server go test ./...

lint:
	$(COMPOSE) run --rm server sh -c 'test -z "$$(gofmt -l .)" || (echo "Go files are not formatted:"; gofmt -l .; exit 1)'
	$(COMPOSE) run --rm server go vet ./...

run:
	$(COMPOSE) up server

stop:
	$(COMPOSE) stop

clean:
	$(COMPOSE) down --remove-orphans

remove: clean
