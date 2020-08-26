.DEFAULT_GOAL=build

include .bingo/Variables.mk

BUILD_FLAGS=CGO_ENABLED=0
GOVVV_FLAGS=$(shell $(GOVVV) -flags -pkg $(shell go list ./buildinfo))

build: $(GOVVV)
	$(BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./...
.PHONY: build

build-pow: $(GOVVV)
	$(BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./cmd/pow
.PHONY: build-pow

build-powd: $(GOVVV)
	$(BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./cmd/powd
.PHONY: build-powd

build-powbench: $(GOVVV)
	$(BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./cmd/powbench
.PHONY: build-powbench

docs-pow:
	rm -rf ./cli-docs/pow
	go run ./cmd/pow/main.go docs ./cli-docs/pow
.PHONY: docs-pow

test:
	go test -short -p 2 -parallel 1 -race -timeout 45m ./... 
.PHONY: test

clean-protos:
	find . -type f -name '*.pb.go' -delete
	find . -type f -name '*pb_test.go' -delete
.PHONY: clean-protos

install-protoc:
	cd buildtools && ./protocInstall.sh
	
PROTOCGENGO=$(shell pwd)/buildtools/protoc-gen-go
protos: install-protoc clean-protos
	PATH=$(PROTOCGENGO):$(PATH) ./scripts/protoc_gen_plugin.bash --proto_path=. --plugin_name=go --plugin_out=. --plugin_opt=plugins=grpc,paths=source_relative
.PHONY: protos

# local is what we run when testing locally.
# This does breaking change detection against our local git repository.
.PHONY: buf-local
buf-local: $(BUF)
	$(BUF) check lint
	# $(BUF) check breaking --against-input '.git#branch=master'

# https is what we run when testing in most CI providers.
# This does breaking change detection against our remote HTTPS git repository.
.PHONY: buf-https
buf-https: $(BUF)
	$(BUF) check lint
	# $(BUF) check breaking --against-input "$(HTTPS_GIT)#branch=master"

# ssh is what we run when testing in CI providers that provide ssh public key authentication.
# This does breaking change detection against our remote HTTPS ssh repository.
# This is especially useful for private repositories.
.PHONY: buf-ssh
buf-ssh: $(BUF)
	$(BUF) check lint
	# $(BUF) check breaking --against-input "$(SSH_GIT)#branch=master"
