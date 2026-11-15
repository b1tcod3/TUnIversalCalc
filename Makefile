.PHONY: all build run test test-race vet lint lint-fix fmt clean

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
BIN ?= bin/tuniversalcalc

all: build

build:
	@mkdir -p bin
	go build -trimpath -ldflags "-X main.version=$(VERSION)" -o $(BIN) ./cmd/tuniversalcalc

run:
	go run ./cmd/tuniversalcalc

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

fmt:
	golangci-lint fmt ./...

clean:
	rm -rf bin coverage.out
