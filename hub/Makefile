export PATH := $(PATH):/usr/local/go/bin

.PHONY: build build-go build-web build-release dev install release-patch release-minor release-major clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GO_LDFLAGS := -X main.version=$(VERSION)

build: build-web build-go

build-go:
	go build -trimpath -ldflags "$(GO_LDFLAGS)" -o bin/couchdev ./cmd/couchdev

build-web:
	cd web && npm install && npm run build

# Produces linux/amd64 and linux/arm64 binaries for attaching to a GitHub release.
build-release: build-web
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -o bin/couchdev-linux-amd64 ./cmd/couchdev
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(GO_LDFLAGS)" -o bin/couchdev-linux-arm64 ./cmd/couchdev

release-patch:
	@bash release.sh patch

release-minor:
	@bash release.sh minor

release-major:
	@bash release.sh major

dev:
	@bash dev.sh

# Bare-metal install using the locally-built binary. Run `make build` first.
install: build
	@bash install.sh --local-bin bin/couchdev

clean:
	rm -rf bin/ .dev-run/ web/dist/ web/node_modules/
