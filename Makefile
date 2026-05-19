.PHONY: build test lint fmt vet all

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o jamshid ./cmd/jamshid

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

all: fmt vet lint test build
