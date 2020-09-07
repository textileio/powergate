localnet-down:
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: localnet-down

localnet: 
	BIGSECTORS=true docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build -V
.PHONY: localnet


up: down
	LOTUS_IMAGE_TAG=v0.6.0 \
	docker-compose \
		-p testnet \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: up

down:
	LOTUS_IMAGE_TAG=v0.6.0 \
	docker-compose \
		-p testnet \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: down

nerpa-up: nerpa-down
	LOTUS_IMAGE_TAG=nerpa-ntwk-nerpa-8.28.0 \
	docker-compose \
		-p nerpa \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: nerpa-up

nerpa-down:
	LOTUS_IMAGE_TAG=nerpa-ntwk-nerpa-8.28.0 \
	docker-compose \
		-p nerpa \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: nerpa-down
