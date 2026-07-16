BINARY_NAME=oc-sync
BINARY_PATH=bin/$(BINARY_NAME)
CMD_PATH=./cmd/oc-sync

GOFLAGS ?=
GO ?= go
GOLANGCI_LINT ?= golangci-lint

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X github.com/dagimg-dot/oc-sync/internal/cli.version=$(VERSION) \
	-X github.com/dagimg-dot/oc-sync/internal/cli.commit=$(COMMIT) \
	-X github.com/dagimg-dot/oc-sync/internal/cli.buildDate=$(BUILD_DATE)

.PHONY: all build clean fmt check-fmt lint lint-fix test test-v vet install coverage

all: fmt build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BINARY_PATH) $(CMD_PATH)

run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run $(GOFLAGS) $(CMD_PATH) $(filter-out $@,$(MAKECMDGOALS))

run-build: build
	@echo "Running $(BINARY_PATH)..."
	./$(BINARY_PATH) $(filter-out $@,$(MAKECMDGOALS))

fmt:
	$(GO) fmt ./...

check-fmt:
	@if [ -n "$$($(GO) fmt ./...)" ]; then \
		echo "Unformatted Go files. Run 'make fmt'."; \
		exit 1; \
	fi

lint:
	$(GOLANGCI_LINT) run ./...

lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

test:
	$(GO) test ./... $(GOFLAGS)

test-v:
	$(GO) test -v ./... $(GOFLAGS)

vet:
	$(GO) vet ./...

install:
	$(GO) install $(GOFLAGS) -ldflags="$(LDFLAGS)" $(CMD_PATH)

coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	@rm -rf bin/ coverage.out coverage.html

tidy:
	$(GO) mod tidy
