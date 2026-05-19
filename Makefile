.PHONY: build test lint fmt vet all

build:
	go build -o jamshid .

test:
	go test ./... -v -race

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

all: fmt vet lint test build
