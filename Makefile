GO ?= go
WAILS ?= wails
WEBKIT_TAG ?= webkit2_41

.PHONY: help build dev test fmt doctor clean

help:
	@printf '%s\n' \
		'build   Build the production binary and embed .env' \
		'dev     Run the app in development mode' \
		'test    Run all Go tests' \
		'fmt     Format Go source files' \
		'doctor  Check Wails system dependencies' \
		'clean   Remove generated build output'

build:
	@test -f .env || (printf '%s\n' 'Error: .env is required for a release build.' >&2; exit 1)
	$(WAILS) build -tags "$(WEBKIT_TAG)"

dev:
	$(WAILS) dev

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

doctor:
	$(WAILS) doctor

clean:
	rm -rf build/bin frontend/dist
