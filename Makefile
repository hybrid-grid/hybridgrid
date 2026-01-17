.PHONY: all build clean test lint proto-gen install

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
GOBIN := $(shell go env GOPATH)/bin

all: build

build: proto-gen
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/hgbuild ./cmd/hgbuild
	go build $(LDFLAGS) -o bin/hg-coord ./cmd/hg-coord
	go build $(LDFLAGS) -o bin/hg-worker ./cmd/hg-worker

proto-gen:
	@echo "Generating protobuf code..."
	@mkdir -p gen/go/hybridgrid/v1
	protoc --go_out=gen/go --go_opt=module=github.com/h3nr1-d14z/hybridgrid/gen/go \
		--go-grpc_out=gen/go --go-grpc_opt=module=github.com/h3nr1-d14z/hybridgrid/gen/go \
		-I proto \
		proto/hybridgrid/v1/build.proto

proto-install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

test:
	go test -v -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	INTEGRATION_TEST=1 go test -v ./test/integration/...

lint:
	golangci-lint run

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

install: build
	sudo cp bin/hgbuild /usr/local/bin/
	sudo cp bin/hg-coord /usr/local/bin/
	sudo cp bin/hg-worker /usr/local/bin/

run-coord:
	go run ./cmd/hg-coord serve

run-worker:
	go run ./cmd/hg-worker serve
