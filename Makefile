NAME = $(shell basename $(PWD))

up: down
	docker compose up -d --build

logs:
	docker compose logs

shell: up
	docker exec -it $(NAME)-alpine-1 /bin/sh

down:
	docker compose down -v --remove-orphans
