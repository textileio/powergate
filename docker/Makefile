down:
	POWD_NETWORK=testnet LOTUS_IMAGE_TAG=v0.4.1 \
	docker-compose \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: down

up: down
	POWD_NETWORK=testnet LOTUS_IMAGE_TAG=v0.4.1 \
	docker-compose \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: up


localnet-down:
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: localnet-down

localnet: 
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build -V
.PHONY: localnet


nerpa-up: nerpa-down
	POWD_NETWORK=nerpa LOTUS_IMAGE_TAG=nerpa-ntwk-nerpa-7.7.0 \
	docker-compose \
		-p nerpa \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: nerpa-up

nerpa-down:
	POWD_NETWORK=nerpa LOTUS_IMAGE_TAG=nerpa-ntwk-nerpa-7.7.0 \
	docker-compose \
		-p nerpa \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: nerpa-down
