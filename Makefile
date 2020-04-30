clean:
	rm -f powd pow powbench
.PHONY: clean

build-cli:
	go build -o pow exe/cli/main.go
.PHONY: build-cli

build-server:
	go build -o powd exe/server/main.go
.PHONY: build-server

build-bench:
	go build -o powbench exe/bench/main.go
.PHONY: build-bench

build: build-cli build-server
.PHONY: build

test:
	go test -short -p 4 -race ./... 
.PHONY: test
