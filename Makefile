.PHONY: build build-go build-web clean

build: build-web build-go

build-go:
	go build -trimpath -o bin/couchdev ./cmd/couchdev

build-web:
	cd web && npm install && npm run build

clean:
	rm -rf bin/ web/dist/ web/node_modules/
