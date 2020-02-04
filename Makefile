SUBMODULES=.update-modules

FFI_PATH:=./extern/filecoin-ffi/
FFI_DEPS:=libfilecoin.a filecoin.pc filecoin.h
FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))

$(FFI_DEPS): .filecoin-build ;

.filecoin-build: $(FFI_PATH)
	$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
	@touch $@

.update-modules:
	git submodule update --init --recursive
	@touch $@

build: .update-modules .filecoin-build
.PHONY: build

clean:
	rm -f .filecoin-build
	rm -f .update-modules

test: build
	mkdir -p /var/tmp/filecoin-proof-parameters
	cat build/proof-params/parameters.json | jq 'keys[]' | xargs touch
	mv -n v20* /var/tmp/filecoin-proof-parameters
	rm v20* || true
	go test -short -p 1 -v -timeout 30s ./... 
.PHONY: test