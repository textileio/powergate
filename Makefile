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
	go test -short -p 4 -race ./... 
.PHONY: test

clean-protos:
	rm -f **/*.pb.go
	rm -f **/*pb_test.go
.PHONY: clean-protos

protos: clean-protos
	./scripts/protoc_gen_plugin.bash --proto_path=. --plugin_name=go --plugin_out=. --plugin_opt=plugins=grpc
.PHONY: protos
