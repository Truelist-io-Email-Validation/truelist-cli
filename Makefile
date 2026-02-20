BINARY     := truelist
MODULE     := github.com/Truelist-io-Email-Validation/truelist-cli
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -s -w -X $(MODULE)/cmd.Version=$(VERSION)
BUILD_DIR  := dist

.PHONY: build install clean test lint fmt vet all

all: fmt vet lint test build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -rf $(BUILD_DIR)

test:
	go test -race -cover ./...

lint:
	@which golangci-lint > /dev/null 2>&1 || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

# Cross-compilation targets
build-all: build-linux build-darwin build-windows

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-windows-amd64.exe .
