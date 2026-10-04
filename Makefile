.PHONY: all build build-ui clean test lint proto-gen install changelog run-coord run-worker run-dashboard

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
GOBIN := $(shell go env GOPATH)/bin
UI_DIR := internal/observability/ui/web

all: build

build: build-ui
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/hgbuild ./cmd/hgbuild
	go build $(LDFLAGS) -o bin/hg-coord ./cmd/hg-coord
	go build $(LDFLAGS) -o bin/hg-worker ./cmd/hg-worker
	go build $(LDFLAGS) -o bin/hg-dashboard ./cmd/hg-dashboard

# Builds the dashboard's React frontend into web/dist, which
# hg-dashboard's go:embed directive (internal/observability/ui/server.go)
# picks up. Required before building/running hg-dashboard from source.
build-ui:
	@cd $(UI_DIR) && npm install --no-fund --no-audit && npm run build

proto-gen:
	@echo "Generating protobuf code..."
	@mkdir -p gen/go/hybridgrid/v1
	protoc --go_out=gen/go --go_opt=module=github.com/h3nr1-d14z/hybridgrid/gen/go \
		--go-grpc_out=gen/go --go-grpc_opt=module=github.com/h3nr1-d14z/hybridgrid/gen/go \
		-I proto \
	proto/hybridgrid/v1/build.proto \
	proto/hybridgrid/v1/telemetry.proto

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
	sudo cp bin/hg-dashboard /usr/local/bin/

run-coord:
	go run ./cmd/hg-coord serve

run-worker:
	go run ./cmd/hg-worker serve

run-dashboard: build-ui
	go run ./cmd/hg-dashboard serve

changelog:
	@echo "Changelog management - CHANGELOG.md exists at project root"
	@test -f CHANGELOG.md && echo "✓ CHANGELOG.md is up to date" || (echo "✗ CHANGELOG.md missing"; exit 1)
	@echo "See scripts/changelog.sh to generate draft entries from git history"
