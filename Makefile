# Makefile

.DEFAULT_GOAL := help

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.11.4

# Default tag prefix. Override with `make patch VERSION_PREFIX=`.
VERSION_PREFIX ?= v

##@ Tagging

LATEST_TAG = $(shell git tag --list "$(VERSION_PREFIX)*" --sort=-v:refname | head -n 1)
VERSION = $(shell [ -n "$(LATEST_TAG)" ] && echo "$(LATEST_TAG)" | sed "s/^$(VERSION_PREFIX)//" || echo "0.0.0")

.PHONY: patch
patch: ## Create a new patch release tag.
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

.PHONY: minor
minor: ## Create a new minor release tag.
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.%d.0", $$1, $$2+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

.PHONY: major
major: ## Create a new major release tag.
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.0.0", $$1+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

.PHONY: tag
tag: ## Show latest tag.
	@echo "Latest version: $(LATEST_TAG)"

.PHONY: push
push: ## Push tags to remote.
	git push --tags

##@ Development

.PHONY: download
download: ## Download Go packages.
	go mod download

.PHONY: run
run: ## Run recall locally.
	go run ./cmd --help

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...
	bash scripts/recall-fzf_test.sh

.PHONY: cover
cover: ## Generate and open coverage report.
	go test -coverprofile=coverage.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

.PHONY: build
build: ## Build recall.
	go build -ldflags="-s -w" -o recall ./cmd

.PHONY: install
install: build ## Install recall and recall-fzf locally.
	mkdir -p ~/.local/bin
	mv recall ~/.local/bin/recall
	chmod +x scripts/recall-fzf
	cp scripts/recall-fzf ~/.local/bin/recall-fzf

.PHONY: clean
clean: ## Clean generated files.
	rm -f recall coverage.out coverage.html

.PHONY: lint
lint: golangci-lint ## Run golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint with fixes.
	$(GOLANGCI_LINT) run --fix

##@ Dependencies

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if needed.

$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3); \
echo "Downloading $${package}"; \
rm -f $(1) || true; \
GOBIN=$(LOCALBIN) go install $${package}; \
mv $(1) $(1)-$(3); \
}; \
ln -sf $(1)-$(3) $(1)
endef

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

