localnet-down:
	DOCKER_BUILDKIT=1 \
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: localnet-down

localnet: 
	DOCKER_BUILDKIT=1 \
	BIGSECTORS=true \
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build -V
.PHONY: localnet


up: down
	DOCKER_BUILDKIT=1 \
	LOTUS_IMAGE_TAG=v1.10.0 \
	docker-compose \
		-p mainnet \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: up

down:
	DOCKER_BUILDKIT=1 \
	LOTUS_IMAGE_TAG=v1.10.0 \
	docker-compose \
		-p mainnet \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: down


up-calibration: down-calibration
	DOCKER_BUILDKIT=1 \
	LOTUS_IMAGE_TAG=ntwk-calibration-d6c42 \
	docker-compose \
		-p calibration \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: up-calibration

down-calibration:
	DOCKER_BUILDKIT=1 \
	LOTUS_IMAGE_TAG=ntwk-calibration-d6c42 \
	docker-compose \
		-p calibration \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: down-calibration
