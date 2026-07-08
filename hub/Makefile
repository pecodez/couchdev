.PHONY: build build-go build-web build-release install clean

build: build-web build-go

build-go:
	go build -trimpath -o bin/couchdev ./cmd/couchdev

build-web:
	cd web && npm install && npm run build

# Produces linux/amd64 and linux/arm64 binaries for attaching to a GitHub release.
build-release: build-web
	GOOS=linux GOARCH=amd64 go build -trimpath -o bin/couchdev-linux-amd64 ./cmd/couchdev
	GOOS=linux GOARCH=arm64 go build -trimpath -o bin/couchdev-linux-arm64 ./cmd/couchdev

# Bare-metal install using the locally-built binary. Run `make build` first.
install: build
	@bash install.sh --local-bin bin/couchdev

clean:
	rm -rf bin/ web/dist/ web/node_modules/
