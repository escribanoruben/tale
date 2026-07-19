BINARY := tale
GO     := go
ARGS   ?= tui

.PHONY: help build run test wordlists stats clean install

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*## "}; {printf "  %-12s %s\n", $$1, $$2}'

build: ## Build the tale binary
	$(GO) build -o $(BINARY) ./cmd/tale

run: ## Run tale without installing (override with ARGS=...)
	$(GO) run ./cmd/tale $(ARGS)

test: ## Run all tests
	$(GO) test ./...

wordlists: ## Regenerate curated word lists (requires network)
	$(GO) run ./cmd/prepare-wordlists

stats: build ## Show embedded dictionary statistics
	./$(BINARY) stats

install: ## Install tale to $$GOPATH/bin
	$(GO) install ./cmd/tale

clean: ## Remove build artifacts
	rm -f $(BINARY)
