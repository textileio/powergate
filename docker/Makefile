LT = $(shell cat ~/.lotus/token)
#LT = $(shell docker exec lotustest cat /root/.lotus/token)

fresh: down up
.PHONY: fresh-up

down:
	@TEXTILE_LOTUS_TOKEN=$(LT) docker-compose down
.PHONY: down

up: 
	TEXTILE_LOTUS_TOKEN=$(LT) docker-compose up --build 
.PHONY: up