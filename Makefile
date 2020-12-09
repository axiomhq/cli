# TOOLCHAIN
GO				:= CGO_ENABLED=0 GOBIN=$(CURDIR)/bin go
GO_BIN_IN_PATH  := CGO_ENABLED=0 go
GOFMT			:= $(GO)fmt

# ENVIRONMENT
VERBOSE		=
GOPATH		:= $(GOPATH)
MOD_NAME	:= github.com/axiomhq/cli

# APPLICATION INFORMATION
BUILD_DATE      := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
REVISION        := $(shell git rev-parse --short HEAD)
RELEASE         := $(shell git describe --tags 2>/dev/null || git rev-parse --short HEAD)-dev
USER            := $(shell whoami)

# GO TOOLS
GEN_CLI_DOCS	:= bin/gen-cli-docs
GOLANGCI_LINT	:= bin/golangci-lint
GORELEASER		:= bin/goreleaser
GOTESTSUM		:= bin/gotestsum

# MISC
COVERPROFILE	:= coverage.out
DIST_DIR		:= dist
MANPAGES_DIR	:= man

# TAGS
GOTAGS := osusergo netgo static_build

# FLAGS
GOFLAGS := -buildmode=exe -tags='$(GOTAGS)' -installsuffix=cgo -trimpath
GOFLAGS += -ldflags='-s -w -extldflags "-fno-PIC -static"
GOFLAGS += -X $(MOD_NAME)/pkg/version.release=$(RELEASE)
GOFLAGS += -X $(MOD_NAME)/pkg/version.revision=$(REVISION)
GOFLAGS += -X $(MOD_NAME)/pkg/version.buildDate=$(BUILD_DATE)
GOFLAGS += -X $(MOD_NAME)/pkg/version.buildUser=$(USER)'

GO_TEST_FLAGS		:= -race -coverprofile=$(COVERPROFILE)
GORELEASER_FLAGS	:= --snapshot --rm-dist

# DEPENDENCIES
GOMODDEPS = go.mod go.sum

# Enable verbose test output if explicitly set.
GOTESTSUM_FLAGS	=
ifdef VERBOSE
	GOTESTSUM_FLAGS += --format=standard-verbose
endif

# FUNCTIONS
# func go-list-pkg-sources(package)
go-list-pkg-sources = $(GO) list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' $(1)
# func go-pkg-sourcefiles(package)
go-pkg-sourcefiles = $(shell $(call go-list-pkg-sources,$(strip $1)))

.PHONY: all
all: dep fmt lint test build man ## Run dep, fmt, lint, test, build and man

.PHONY: build
build: $(GORELEASER) dep.stamp $(call go-pkg-sourcefiles, ./...) ## Build the binaries
	@echo ">> building binaries"
	@$(GORELEASER) build $(GORELEASER_FLAGS)

.PHONY: clean
clean: ## Remove build and test artifacts
	@echo ">> cleaning up artifacts"
	@rm -rf $(DIST_DIR) $(COVERPROFILE)

.PHONY: cover
cover: $(COVERPROFILE) ## Calculate the code coverage score
	@echo ">> calculating code coverage"
	@$(GO) tool cover -func=$(COVERPROFILE) | tail -n1

.PHONY: dep-clean
dep-clean: ## Remove obsolete dependencies
	@echo ">> cleaning dependencies"
	@$(GO) mod tidy

.PHONY: dep-upgrade
dep-upgrade: ## Upgrade all direct dependencies to their latest version
	@echo ">> upgrading dependencies"
	@$(GO) get -d $(shell $(GO) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@make dep

.PHONY: dep
dep: dep-clean dep.stamp ## Install and verify dependencies and remove obsolete ones

dep.stamp: $(GOMODDEPS)
	@echo ">> installing dependencies"
	@$(GO) mod download
	@$(GO) mod verify
	@touch $@

.PHONY: fmt
fmt: ## Format and simplify the source code using `gofmt`
	@echo ">> formatting code"
	@! $(GOFMT) -s -w $(shell find . -path -prune -o -name '*.go' -print) | grep '^'

.PHONY: install
install: $(GOPATH)/bin/axiom ## Install the binary into the $GOPATH/bin directory

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Lint the source code
	@echo ">> linting code"
	@$(GOLANGCI_LINT) run

.PHONY: man
man: $(GEN_CLI_DOCS) ## Generate man pages
	@echo ">> generate man pages"
	@rm -rf $(MANPAGES_DIR)
	@$(GEN_CLI_DOCS) -d=$(MANPAGES_DIR) -t=$(RELEASE)

.PHONY: test
test: $(GOTESTSUM) ## Run all tests. Run with VERBOSE=1 to get verbose test output (`-v` flag)
	@echo ">> running tests"
	@$(GOTESTSUM) $(GOTESTSUM_FLAGS) -- $(GO_TEST_FLAGS) ./...

.PHONY: tools
tools: $(GEN_CLI_DOCS) $(GOLANGCI_LINT) $(GORELEASER) $(GOTESTSUM) ## Install all tools into the projects local $GOBIN directory

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# MISC TARGETS

$(COVERPROFILE):
	@make test

# INSTALL TARGETS

$(GOPATH)/bin/axiom: dep.stamp $(call go-pkg-sourcefiles, ./...)
	@echo ">> installing axiom binary"
	@$(GO_BIN_IN_PATH) install $(GOFLAGS) ./cmd/axiom

# GO TOOLS

$(GEN_CLI_DOCS): dep.stamp $(call go-pkg-sourcefiles, github.com/axiomhq/cli/tools/gen-cli-docs) $(call go-pkg-sourcefiles, github.com/axiomhq/cli/cmd/axiom/...)
	@echo ">> installing gen-cli-docs"
	@$(GO) install github.com/axiomhq/cli/tools/gen-cli-docs

$(GOLANGCI_LINT): dep.stamp $(call go-pkg-sourcefiles, github.com/golangci/golangci-lint/cmd/golangci-lint)
	@echo ">> installing golangci-lint"
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint

$(GORELEASER): dep.stamp $(call go-pkg-sourcefiles, github.com/goreleaser/goreleaser)
	@echo ">> installing goreleaser"
	@$(GO) install github.com/goreleaser/goreleaser

$(GOTESTSUM): dep.stamp $(call go-pkg-sourcefiles, gotest.tools/gotestsum)
	@echo ">> installing gotestsum"
	@$(GO) install gotest.tools/gotestsum
