MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCAL_BIN_PATH := ${PROJECT_PATH}/bin

LINT_GOGC := 10
LINT_TIMEOUT := 10m

## Tools
GOIMPORT ?= $(LOCALBIN)/goimports
GOIMPORT_VERSION ?= latest
GOLANGCI ?= $(LOCALBIN)/golangci-lint
GOLANGCI_VERSION ?= v1.60.1
YQ ?= $(LOCALBIN)/yq
KUBECTL ?= kubectl
KO ?= $(LOCALBIN)/ko
KO_VERSION ?= latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


ifndef ignore-not-found
  ignore-not-found = false
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

.PHONY: build
build: $(LOCALBIN)
	go build -o $(LOCALBIN)/kfm-transform main.go

.PHONY: clean
clean:
	go clean -x
	go clean -x -testcache

.PHONY: fmt
fmt: goimport
	$(GOIMPORT) -l -w .
	go fmt ./...

.PHONY: test
test:
	go test -v ./...

.PHONY: deps
deps:
	go mod tidy

.PHONY: publish
publish: ko ## Deploy test app.
	KO_DOCKER_REPO=quay.io/lburgazzoli $(LOCALBIN)/ko build --platform=linux/amd64,linux/arm64 -B .

.PHONY: check/lint
check: check/lint

.PHONY: check/lint
check/lint: golangci-lint
	@$(GOLANGCI) run \
		--config .golangci.yml \
		--out-format tab \
		--exclude-dirs etc \
		--timeout $(LINT_TIMEOUT)

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	@mkdir -p $(LOCALBIN)

.PHONY: goimport
goimport: $(GOIMPORT)
$(GOIMPORT): $(LOCALBIN)
	@test -s $(GOIMPORT) || \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORT_VERSION)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI)
$(GOLANGCI): $(LOCALBIN)
	@test -s $(GOLANGCI) || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)

.PHONY: yq
yq: $(YQ)
$(YQ): $(LOCALBIN)
	@test -s $(LOCALBIN)/yq || \
	GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@latest

.PHONY: ko
ko: $(KO)
$(KO): $(LOCALBIN)
	@test -s $(LOCALBIN)/ko || \
	GOBIN=$(LOCALBIN) go install github.com/google/ko@$(KO_VERSION)