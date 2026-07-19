BINARY := tale
GO     := go
ARGS   ?= tui

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo unknown)
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

.PHONY: help build run test wordlists stats clean install

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the tale binary
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/tale

run: ## Run tale without installing (override with ARGS=...)
	$(GO) run ./cmd/tale $(ARGS)

test: ## Run all tests
	$(GO) test ./...

wordlists: ## Regenerate curated word lists (requires network)
	$(GO) run ./cmd/prepare-wordlists

stats: build ## Show embedded dictionary statistics
	./$(BINARY) stats

install: ## Install tale to $$GOPATH/bin
	$(GO) install -ldflags "$(LDFLAGS)" ./cmd/tale

clean: ## Remove build artifacts
	rm -f $(BINARY)
