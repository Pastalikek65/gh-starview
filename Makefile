VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
LDFLAGS := -s -w -X main.version=$(VERSION)
BUILD_FLAGS := -trimpath

.PHONY: test vet lint cover build clean help

test:
	go test ./... -count=1 -timeout 30s -cover

vet:
	go vet ./...

lint:
	go vet ./...
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, vet passed"

cover:
	go test ./... -count=1 -timeout 30s -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -20
	@echo "total: $$(go tool cover -func=coverage.out | grep total | awk '{print $$3}')"

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags="$(LDFLAGS)" -o gh-starview .

clean:
	rm -f gh-starview gh-starview-linux-* gh-starview-android-* coverage.out

help:
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'
