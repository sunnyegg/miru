GO ?= go
WAILS ?= wails
WEBKIT_TAG ?= webkit2_41
GOLANGCI_LINT_VERSION ?= v2.13.1
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo dev)
LD_FLAGS ?= -X main.version=$(VERSION)

.PHONY: help deps build dev test check-windows fmt fmt-fe lint lint-fe typecheck doctor clean bench-idle-ram

help:
	@printf '%s\n' \
		'deps      Install Go modules and frontend npm packages' \
		'build     Build the production binary and embed .env' \
		'dev       Run the app in development mode' \
		'test      Run all Go tests' \
		'check-windows  Cross-compile and vet for Windows (GOOS=windows)' \
		'fmt       Format Go source files' \
		'fmt-fe    Format frontend source files' \
		'lint      Run golangci-lint (standard + nestif)' \
		'lint-fe   Run ESLint on the frontend' \
		'typecheck Run TypeScript type checking on the frontend (frontend/)' \
		'doctor    Check Wails system dependencies' \
		'clean     Remove generated build output' \
		'bench-idle-ram  Measure idle RAM on Linux (requires built binary)'

deps:
	$(GO) mod download
	NODE_ENV=development npm --prefix frontend install

build:
	@test -f .env || (printf '%s\n' 'Error: .env is required for a release build.' >&2; exit 1)
	$(WAILS) build -tags "$(WEBKIT_TAG)" -ldflags "$(LD_FLAGS)"

dev:
	$(WAILS) dev -tags "$(WEBKIT_TAG)" -ldflags "$(LD_FLAGS)"

test:
	$(GO) test ./...

check-windows:
	GOOS=windows GOARCH=amd64 $(GO) build ./...
	GOOS=windows GOARCH=amd64 $(GO) vet ./...

fmt:
	$(GO) fmt ./...

fmt-fe:
	npm --prefix frontend run format

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run

lint-fe:
	npm --prefix frontend run lint

typecheck:
	npm --prefix frontend run typecheck

doctor:
	$(WAILS) doctor

clean:
	rm -rf build/bin frontend/dist

bench-idle-ram:
	@chmod +x scripts/bench-idle-ram.sh
	@scripts/bench-idle-ram.sh
