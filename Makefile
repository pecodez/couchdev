export PATH := $(PATH):/home/phil/go-sdk/go/bin

HUB     := hub
WEB     := hub/web

.PHONY: test test-go test-unit test-vrt

## test-go: Go unit + integration tests
test-go:
	cd $(HUB) && go test ./...

## test-unit: Frontend unit tests (Vitest)
test-unit:
	cd $(WEB) && npm test

## test-vrt: Playwright visual regression tests against Storybook
test-vrt:
	cd $(WEB) && npm run vrt

## test: full test suite
test: test-go test-unit test-vrt
