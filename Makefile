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
	go test -short -p 2 -parallel 2 -race -timeout 45m ./... 
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
