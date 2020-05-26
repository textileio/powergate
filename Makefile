clean:
	rm -f powd pow powbench
.PHONY: clean

BUILD_FLAGS=CGO_ENABLED=0
build-cli:
	$(BUILD_FLAGS) go build -o pow exe/cli/main.go
.PHONY: build-cli

build-server:
	$(BUILD_FLAGS) go build -o powd exe/server/main.go
.PHONY: build-server

build-bench:
	$(BUILD_FLAGS) go build -o powbench exe/bench/main.go
.PHONY: build-bench

build: build-cli build-server build-bench
.PHONY: build

test:
	go test -short -p 4 -race -timeout 30m ./... 
.PHONY: test

clean-protos:
	find . -type f -name '*.pb.go' -delete
	find . -type f -name '*pb_test.go' -delete
.PHONY: clean-protos

install-protoc:
	cd buildtools && ./protocInstall.sh
	
PROTOCBIN=$(pwd)/buildtools/protoc/bin
PROTOCGENGO=$(pwd)/buildtools/protoc-gen-go
BINARIES=$(PROTOCBIN):$(PROTOCGENGO)
protos: install-protoc clean-protos
	PATH=$(BINARIES):$(PATH) ./scripts/protoc_gen_plugin.bash --proto_path=. --plugin_name=go --plugin_out=. --plugin_opt=plugins=grpc
.PHONY: protos
