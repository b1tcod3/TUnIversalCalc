.PHONY: all build run test test-race vet fmt clean

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

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

clean:
	rm -rf bin coverage.out
