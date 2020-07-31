COMPAT_MSG="The current version of Powergate support newer versions than Testnet network. You might be interested in using the Nerpa network (make nerpa-up)"

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
	docker-compose -p localnet \
		-f docker-compose-localnet.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build -V
.PHONY: localnet


nerpa-up: nerpa-down
	LOTUS_IMAGE_TAG=nerpa-ntwk-nerpa-7.7.0 \
	docker-compose \
		-p nerpa \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		up --build 
.PHONY: nerpa-up

nerpa-down:
	LOTUS_IMAGE_TAG=nerpa-ntwk-nerpa-7.7.0 \
	docker-compose \
		-p nerpa \
		-f docker-compose.yaml \
		-f ipfs-image.yaml \
		-f powergate-build-context.yaml \
		down
.PHONY: nerpa-down
