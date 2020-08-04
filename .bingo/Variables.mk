# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.2.2. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Bellow generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for govvv variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(GOVVV)
#	@echo "Running govvv"
#	@$(GOVVV) <flags/args..>
#
GOVVV := $(GOBIN)/govvv-v0.3.0
$(GOVVV): .bingo/govvv.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/govvv-v0.3.0"
	@cd .bingo && $(GO) build -modfile=govvv.mod -o=$(GOBIN)/govvv-v0.3.0 "github.com/ahmetb/govvv"

GOX := $(GOBIN)/gox-v1.0.1
$(GOX): .bingo/gox.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/gox-v1.0.1"
	@cd .bingo && $(GO) build -modfile=gox.mod -o=$(GOBIN)/gox-v1.0.1 "github.com/mitchellh/gox"

