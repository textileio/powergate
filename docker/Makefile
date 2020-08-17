COMPAT_MSG="The current version of Powergate support newer versions than Testnet network. You might be interested in using the Calibration network (make calibration-up)"

down:
	@echo ${COMPAT_MSG}
.PHONY: down

up: down
.PHONY: up


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


calibration-up: calibration-down
	LOTUS_IMAGE_TAG=ntwk-calibration-8.13.1-f094273f \
	docker-compose \
		-p calib \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: calibration-up

calibration-down:
	LOTUS_IMAGE_TAG=ntwk-calibration-8.13.1-f094273f \
	docker-compose \
		-p calib \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: calibration-down
