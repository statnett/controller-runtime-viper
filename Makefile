# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: test

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

GO_SOURCE_FILES = $(shell find . -type f -name '*.go' -not -name 'zz_generated*')
LOCAL_GO_MODULE = $(shell head -n 1 go.mod | awk '{print $$2}')
fmt-imports: goimports ## Format imports in code using goimports.
	# Between patterns 'import (' and ')', delete lines matching '^[[:space:]]*$' (lines only containing whitespace)
	sed -i -e '/import (/,/)/{/^[[:space:]]*$$/d}' ${GO_SOURCE_FILES}
	$(GOIMPORTS) -w --local ${LOCAL_GO_MODULE} ${GO_SOURCE_FILES}

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: vet ## Run tests.
	go test -v -race ./... -coverprofile cover.out

.PHONY: go-mod-tidy
go-mod-tidy: ## Run go mod tidy against code.
	go mod tidy

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOIMPORTS ?= $(LOCALBIN)/goimports

## Tool Versions
GOIMPORTS_VERSION ?= v0.5.0

.PHONY: goimports
goimports: $(GOIMPORTS) ## Download goimports locally if necessary.
$(GOIMPORTS):
	test -s $(LOCALBIN)/goimports || GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
