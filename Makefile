clean:
	rm -f powd pow
.PHONY: clean

build-cli:
	go build -o pow exe/cli/main.go
.PHONY: build-cli

build-server:
	go build -o powd exe/server/main.go
.PHONY: build-server

build: build-cli build-server
.PHONY: build

test:
	go test -short -p 1 ./... 
.PHONY: test
