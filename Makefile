.DEFAULT_GOAL=build

include .bingo/Variables.mk

BUILD_FLAGS=CGO_ENABLED=0
POW_VERSION ?= "none"
GOVVV_FLAGS=$(shell $(GOVVV) -flags -version $(POW_VERSION) -pkg $(shell go list ./buildinfo))

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

define gen_release_files
	$(GOX) -osarch=$(3) -output="build/$(2)/$(2)_${POW_VERSION}_{{.OS}}-{{.Arch}}/$(2)" -ldflags="${GOVVV_FLAGS}" $(1)
	mkdir -p build/dist; \
	cd build/$(2); \
	for release in *; do \
		cp ../../LICENSE ../../README.md $${release}/; \
		if [[ $${release} != *"windows"* ]]; then \
  		POW_FILE=$(2) $(GOMPLATE) -f ../../dist/install.tmpl -o "$${release}/install"; \
			tar -czvf ../dist/$${release}.tar.gz $${release}; \
		else \
			zip -r ../dist/$${release}.zip $${release}; \
		fi; \
	done
endef

build-pow-release: $(GOX) $(GOVVV) $(GOMPLATE)
	$(call gen_release_files,./cmd/pow,pow,"linux/amd64 linux/386 linux/arm darwin/amd64 windows/amd64")
.PHONY: build-release

build-powd-release: $(GOX) $(GOVVV) $(GOMPLATE)
	$(call gen_release_files,./cmd/powd,powd,"linux/amd64 darwin/amd64")
.PHONY: build-release

build-powbench-release: $(GOX) $(GOVVV) $(GOMPLATE)
	$(call gen_release_files,./cmd/powbench,powbench,"linux/amd64 linux/386 linux/arm darwin/amd64 windows/amd64")
.PHONY: build-release

build-releases: build-pow-release build-powd-release build-powbench-release
.PHONY: build-releases

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
