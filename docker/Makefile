LT = $(shell cat ~/.lotus/token)
#LT = $(shell docker exec lotustest cat /root/.lotus/token)

fresh: down up
.PHONY: fresh-up

down:
	@TEXFILTOOLS_LOTUSTOKEN=$(LT) docker-compose down
.PHONY: down

up: 
	TEXFILTOOLS_LOTUSTOKEN=$(LT) docker-compose up --build 
.PHONY: up