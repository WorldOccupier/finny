.PHONY: install dev build preview

COMPOSE_FILE ?= ../docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

install:
	$(COMPOSE) run --rm web npm install

dev:
	$(COMPOSE) up web server

build:
	$(COMPOSE) run --rm web npm run build

preview:
	$(COMPOSE) run --rm --service-ports web npm run preview
