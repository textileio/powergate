down:
	docker-compose down
	docker-compose -p devnet down
.PHONY: down

up: down
	docker-compose up --build 
.PHONY: up

devnet: 
	docker-compose -p devnet -f docker-compose-devnet.yaml up --build -V
.PHONY: devnet

devnet-small: 
	BIGSECTORS=false docker-compose -p devnet -f docker-compose-devnet.yaml up --build -V
.PHONY: devnet-small
