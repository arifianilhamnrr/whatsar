BINARY     := whatsar
CLI_BINARY := whatsar-cli
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X main.version=$(VERSION)

GO      := go
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)

.PHONY: all build build-cli run run-cli test clean deps release

all: build build-cli

deps:
	$(GO) mod download
	$(GO) mod tidy

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/server

build-cli:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(CLI_BINARY) ./cmd/cli

run:
	$(GO) run ./cmd/server

run-cli:
	$(GO) run ./cmd/cli

test:
	$(GO) test ./...

clean:
	rm -f $(BINARY) $(BINARY).exe

# Cross-compile targets
build-linux-amd64:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_linux_amd64 ./cmd/server

build-linux-arm64:
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_linux_arm64 ./cmd/server

build-linux-armv7:
	GOOS=linux GOARCH=arm GOARM=7 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_linux_armv7 ./cmd/server

build-windows-amd64:
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)_windows_amd64.exe ./cmd/server

release: build-linux-amd64 build-linux-arm64 build-linux-armv7 build-windows-amd64
	@echo "Release binaries built in dist/"