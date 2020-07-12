BUILD_FLAGS=CGO_ENABLED=0

build: 
	$(BUILD_FLAGS) go install ./...
.PHONY: build

build-pow:
	$(BUILD_FLAGS) go install ./cmd/pow
.PHONY: build-pow

build-powd:
	$(BUILD_FLAGS) go install ./cmd/powd
.PHONY: build-powd

build-powbench:
	$(BUILD_FLAGS) go install ./cmd/powbench
.PHONY: build-powbench

docs-pow:
	rm -rf ./cli-docs/pow
	go run ./cmd/pow/main.go docs ./cli-docs/pow
.PHONY: docs-pow

test:
	go test -short -parallel 6 -race -timeout 30m ./... 
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
