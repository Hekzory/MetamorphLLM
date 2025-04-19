# Makefile for MetamorphLLM project

BUILDDIR := build
BINARIES := rewriter suspicious manager

# Check if .env file exists and include it
ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: all clean test build run-suspicious run-rewriter run-manager build-all env-check

all: build-all

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

build-all: $(BUILDDIR) $(BINARIES)

rewriter: $(BUILDDIR)
	go build -o $(BUILDDIR)/rewriter ./cmd/rewriter

suspicious: $(BUILDDIR)
	go build -o $(BUILDDIR)/suspicious ./cmd/suspicious

manager: $(BUILDDIR)
	go build -o $(BUILDDIR)/manager ./cmd/manager

test:
	go test ./internal/...

clean:
	rm -rf $(BUILDDIR)

# Check if GEMINI_API_KEY is set
env-check:
	@if [ -z "$(GEMINI_API_KEY)" ]; then \
		echo "Error: GEMINI_API_KEY is not set. Please set it in your .env file or environment."; \
		echo "You can copy env.example to .env and add your key there."; \
		exit 1; \
	fi

run-suspicious: suspicious
	$(BUILDDIR)/suspicious

run-rewriter: rewriter env-check
	$(BUILDDIR)/rewriter

run-manager: build-all env-check
	$(BUILDDIR)/manager -rewriter $(BUILDDIR)/rewriter -suspicious internal/suspicious/suspicious.go

run-manager-force: build-all env-check
	$(BUILDDIR)/manager -rewriter $(BUILDDIR)/rewriter -suspicious internal/suspicious/suspicious.go -force-rewrite

run-manager-dry: build-all env-check
	$(BUILDDIR)/manager -rewriter $(BUILDDIR)/rewriter -suspicious internal/suspicious/suspicious.go -dry-run

setup-env:
	@if [ ! -f .env ]; then \
		cp env.example .env; \
		echo ".env file created. Please edit it with your API keys."; \
	else \
		echo ".env file already exists."; \
	fi

help:
	@echo "Available targets:"
	@echo "  make build-all    - Build all binaries"
	@echo "  make test         - Run all tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make setup-env    - Create .env file from template if it doesn't exist"
	@echo "  make run-suspicious - Build and run the suspicious program"
	@echo "  make run-rewriter - Build and run the rewriter program"
	@echo "  make run-manager  - Build all and run the manager program"
	@echo "  make run-manager-force - Run manager with force-rewrite enabled"
	@echo "  make run-manager-dry - Run manager in dry-run mode (no deployment)" 