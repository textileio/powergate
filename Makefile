.DEFAULT_GOAL=install

include .bingo/Variables.mk

POW_BUILD_FLAGS?=CGO_ENABLED=0
POW_VERSION?="git"
GOVVV_FLAGS=$(shell $(GOVVV) -flags -version $(POW_VERSION) -pkg $(shell go list ./buildinfo))

build: $(GOVVV)
	$(POW_BUILD_FLAGS) go build -ldflags="${GOVVV_FLAGS}" ./...
.PHONY: build

build-pow: $(GOVVV)
	$(POW_BUILD_FLAGS) go build -ldflags="${GOVVV_FLAGS}" ./cmd/pow
.PHONY: build-pow

build-powd: $(GOVVV)
	$(POW_BUILD_FLAGS) go build -ldflags="${GOVVV_FLAGS}" ./cmd/powd
.PHONY: build-powd

build-powbench: $(GOVVV)
	$(POW_BUILD_FLAGS) go build -ldflags="${GOVVV_FLAGS}" ./cmd/powbench
.PHONY: build-powbench

install: $(GOVVV)
	$(POW_BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./...
.PHONY: install

install-pow: $(GOVVV)
	$(POW_BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./cmd/pow
.PHONY: install-pow

install-powd: $(GOVVV)
	$(POW_BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./cmd/powd
.PHONY: install-powd

install-powbench: $(GOVVV)
	$(POW_BUILD_FLAGS) go install -ldflags="${GOVVV_FLAGS}" ./cmd/powbench
.PHONY: install-powbench

define gen_release_files
	$(GOX) -osarch=$(3) -output="build/$(2)/$(2)_${POW_VERSION}_{{.OS}}-{{.Arch}}/$(2)" -ldflags="${GOVVV_FLAGS}" $(1)
	mkdir -p build/dist; \
	cd build/$(2); \
	for release in *; do \
		cp ../../LICENSE ../../README.md $${release}/; \
		if [ $${release} != *"windows"* ]; then \
  		POW_FILE=$(2) $(GOMPLATE) -f ../../dist/install.tmpl -o "$${release}/install"; \
			tar -czvf ../dist/$${release}.tar.gz $${release}; \
		else \
			zip -r ../dist/$${release}.zip $${release}; \
		fi; \
	done
endef

build-pow-release: $(GOX) $(GOVVV) $(GOMPLATE)
	$(call gen_release_files,./cmd/pow,pow,"linux/amd64 darwin/amd64 windows/amd64")
.PHONY: build-pow-release

build-powd-release: $(GOX) $(GOVVV) $(GOMPLATE)
	$(call gen_release_files,./cmd/powd,powd,"linux/amd64 darwin/amd64")
.PHONY: build-powd-release

build-powbench-release: $(GOX) $(GOVVV) $(GOMPLATE)
	$(call gen_release_files,./cmd/powbench,powbench,"linux/amd64 linux/386 linux/arm darwin/amd64 windows/amd64")
.PHONY: build-powbench-release

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

protos: $(BUF) $(PROTOC_GEN_GO) $(PROTOC_GEN_GO_GRPC) clean-protos
	$(BUF) generate --template '{"version":"v1beta1","plugins":[{"name":"go","out":"api/gen","opt":"paths=source_relative","path":$(PROTOC_GEN_GO)},{"name":"go-grpc","out":"api/gen","opt":"paths=source_relative","path":$(PROTOC_GEN_GO_GRPC)}]}'
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
