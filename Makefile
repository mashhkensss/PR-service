SHELL := /bin/bash
GOCMD ?= go
BINARY ?= bin/reviewer-service

.PHONY: run build test lint bench cover compose-up compose-down migrate-up migrate-down

run:
	$(GOCMD) run ./cmd/reviewer-service

build:
	mkdir -p $(dir $(BINARY))
	CGO_ENABLED=0 $(GOCMD) build -o $(BINARY) ./cmd/reviewer-service

test:
	$(GOCMD) test ./...

lint:
	golangci-lint run ./...

bench:
	$(GOCMD) test ./... -bench=. -benchmem

cover:
	$(GOCMD) test ./... -coverprofile=coverage.out
	$(GOCMD) tool cover -func=coverage.out

compose-up:
	docker compose up --build

compose-down:
	docker compose down

migrate-up:
	docker compose run --rm migrate up

migrate-down:
	docker compose run --rm migrate down
