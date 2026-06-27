# Day 21 — Makefile for Docker workflow
# Run: make help

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
IMAGE_NAME  = shop-api

.PHONY: build run test docker-build docker-run compose-up compose-down help

## build: compile the binary locally
build:
	CGO_ENABLED=0 go build \
	  -ldflags="-s -w -X main.Version=$(VERSION) -X main.GitCommit=$(GIT_COMMIT) -X main.BuildTime=$(BUILD_TIME)" \
	  -o bin/server ./cmd/server

## run: build and run locally
run: build
	./bin/server

## test: run tests
test:
	go test ./...

## docker-build: build the Docker image with version metadata
docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg GIT_COMMIT=$(GIT_COMMIT) \
	  --build-arg BUILD_TIME=$(BUILD_TIME) \
	  -t $(IMAGE_NAME):$(VERSION) \
	  -t $(IMAGE_NAME):latest \
	  .

## docker-run: run the image locally
docker-run:
	docker run --rm -p 8080:8080 $(IMAGE_NAME):latest

## compose-up: start the full stack
compose-up:
	docker compose up -d
	@echo "API available at http://localhost:8080"
	@echo "Watching logs (Ctrl-C to stop watching, services keep running):"
	docker compose logs -f api

## compose-down: stop all services
compose-down:
	docker compose down

## compose-clean: stop and delete all data
compose-clean:
	docker compose down -v

## help: print this message
help:
	@grep -E '^## ' Makefile | sed 's/## //'
