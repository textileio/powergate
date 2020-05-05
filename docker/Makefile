down:
	docker-compose down
	docker-compose -p embed down
.PHONY: down

up: down
	docker-compose up --build 
.PHONY: up

embed: 
	docker-compose -p embed -f docker-compose-embedded.yaml up --build -V
.PHONY: embed
