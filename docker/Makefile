LT = $(shell cat ~/.lotus/token)

fresh: down up
.PHONY: fresh-up

down:
	@TEXTILE_LOTUS_TOKEN=$(LT) docker-compose down
.PHONY: down

up: 
	TEXTILE_LOTUS_TOKEN=$(LT) docker-compose up --build 
.PHONY: up