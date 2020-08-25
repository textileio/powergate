localnet-down:
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: localnet-down

localnet: 
	BIGSECTORS=false docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build -V
.PHONY: localnet


up: down
	LOTUS_IMAGE_TAG=testnet-v0.5.3 \
	docker-compose \
		-p testnet \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: up

down:
	LOTUS_IMAGE_TAG=testnet-v0.5.3 \
	docker-compose \
		-p testnet \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: down
