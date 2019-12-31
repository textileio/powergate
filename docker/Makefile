fresh: down up
.PHONY: fresh-up

down:
	@docker-compose down
.PHONY: down

up: 
	TEXTILE_LOTUS_TOKEN=$(shell cat ~/.lotus/token) docker-compose up --build 