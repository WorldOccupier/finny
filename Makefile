.PHONY: restart

restart:
	docker compose up -d --build --force-recreate web server
