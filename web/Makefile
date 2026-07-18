.PHONY: install dev build preview stop clean remove

COMPOSE_FILE ?= ../docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)
export DOCKER_BUILDKIT := 1
export COMPOSE_DOCKER_CLI_BUILD := 1

install:
	$(COMPOSE) run --rm web npm install

dev:
	$(COMPOSE) up web server

build:
	$(COMPOSE) run --rm web npm run build

preview:
	$(COMPOSE) run --rm --service-ports web npm run preview

stop:
	$(COMPOSE) stop

clean:
	$(COMPOSE) down --remove-orphans

remove: clean
