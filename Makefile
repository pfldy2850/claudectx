BINARY := claudectx
BUILD_DIR := bin
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X github.com/pfldy2850/claudectx/internal/cli.Version=$(VERSION)"

.PHONY: build test test-cover lint clean install

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/claudectx

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	@which golangci-lint > /dev/null 2>&1 || echo "golangci-lint not installed"
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out

install: build
	cp $(BUILD_DIR)/$(BINARY) $(GOPATH)/bin/$(BINARY) 2>/dev/null || \
	cp $(BUILD_DIR)/$(BINARY) $(HOME)/go/bin/$(BINARY)
