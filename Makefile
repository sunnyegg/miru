GO ?= go
WAILS ?= wails
WEBKIT_TAG ?= webkit2_41
GOLANGCI_LINT_VERSION ?= v2.13.1

.PHONY: help deps build dev test fmt lint doctor clean

help:
	@printf '%s\n' \
		'deps    Install Go modules and frontend npm packages' \
		'build   Build the production binary and embed .env' \
		'dev     Run the app in development mode' \
		'test    Run all Go tests' \
		'fmt     Format Go source files' \
		'lint    Run golangci-lint (standard + nestif)' \
		'doctor  Check Wails system dependencies' \
		'clean   Remove generated build output'

deps:
	$(GO) mod download
	npm --prefix frontend install

build:
	@test -f .env || (printf '%s\n' 'Error: .env is required for a release build.' >&2; exit 1)
	$(WAILS) build -tags "$(WEBKIT_TAG)"

dev:
	$(WAILS) dev -tags "$(WEBKIT_TAG)"

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

doctor:
	$(WAILS) doctor

clean:
	rm -rf build/bin frontend/dist
