# Day 22 — Production Makefile for Go service
# Usage: make help

# Version metadata from git
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
IMAGE_NAME  = shop-api

# LDFLAGS: strip debug symbols + inject version metadata
LDFLAGS = -s -w \
  -X main.Version=$(VERSION) \
  -X main.GitCommit=$(GIT_COMMIT) \
  -X main.BuildTime=$(BUILD_TIME)

.PHONY: all lint vet test build docker-build ci clean help

## all: run lint, test, and build
all: lint test build

## vet: run go vet (fast, built-in static analysis)
vet:
	go vet ./...

## lint: run go vet + golangci-lint
lint: vet
	@which golangci-lint > /dev/null || (echo "golangci-lint not found: go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1" && exit 1)
	golangci-lint run ./...

## test: run tests with race detector and coverage
test:
	go test -race -coverprofile=coverage.out -covermode=atomic -timeout=120s ./...
	@echo ""
	@echo "Coverage summary:"
	go tool cover -func=coverage.out | tail -1

## test-verbose: run tests with verbose output
test-verbose:
	go test -race -v -timeout=120s ./...

## build: compile the binary
build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o bin/server ./cmd/server
	@echo "Built bin/server (version=$(VERSION) commit=$(GIT_COMMIT))"

## build-all: cross-compile for all target platforms
build-all:
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o bin/server-linux-amd64 ./cmd/server
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/server-linux-arm64 ./cmd/server
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o bin/server-darwin-arm64 ./cmd/server
	@echo "Cross-compiled for linux/amd64, linux/arm64, darwin/arm64"

## docker-build: build the Docker image
docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg GIT_COMMIT=$(GIT_COMMIT) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  -t $(IMAGE_NAME):$(VERSION) \
	  -t $(IMAGE_NAME):latest \
	  .
	@echo "Image: $(IMAGE_NAME):$(VERSION)"
	@docker image ls $(IMAGE_NAME):$(VERSION)

## ci: run the full CI pipeline locally (same as GitHub Actions)
# Use this to reproduce CI failures locally before pushing.
ci: lint test build
	@echo ""
	@echo "CI pipeline passed locally."

## clean: remove build artifacts
clean:
	rm -rf bin/ coverage.out

## help: print this message
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
