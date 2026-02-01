.PHONY: lint lint-fix build run test test-contracts validate-fixtures clean build-all npm-pack

REVIVE = $(shell go env GOPATH)/bin/revive
AIR = $(shell go env GOPATH)/bin/air
GCL_CUSTOM = ./bin/custom-gcl

lint:
	golangci-lint run ./...
	-@$(REVIVE) -config .revive.toml ./... 
	-@$(GCL_CUSTOM) run -c .golangci.custom.yaml ./... 

lint-fix:
	golangci-lint run --fix ./...

build:
	go build -o agent-telegram .

lint-tools:
	golangci-lint custom

run:
	go run .

test:
	go test -v ./...

test-contracts: ## Run contract tests with fixtures
	go test -v -tags=contracts ./...

validate-fixtures: ## Validate fixture files structure
	@echo "Validating fixtures..."
	@for f in testdata/fixtures/**/*.json; do \
		jq empty "$$f" 2>/dev/null || echo "Invalid JSON: $$f"; \
	done
	@echo "Fixtures validated"

record-fixture: ## Record a fixture (use: make record-fixture METHOD=messages.getHistory PEER=@username)
	go run ./testdata/recorder -method $(METHOD) -peer $(PEER) -output ./testdata/fixtures

sanitize-fixture: ## Sanitize a fixture (use: make sanitize-fixture INPUT=path/to/fixture.json)
	go run ./testdata/sanitizer -input $(INPUT) -inplace

clean:
	go clean
	rm -f agent-telegram

login-mock:
	go run main.go login --mock

dev: ## Run with live reload (air) for serve command
	$(AIR) -- -serve

dev-args: ## Run with live reload (air) with custom args
	$(AIR) -- $(ARGS)

install-air: ## Install air for live reloading
	go install github.com/air-verse/air@latest

build-all: ## Build binaries for all platforms
	./scripts/build-all.sh

npm-pack: build-all ## Pack npm package (builds all platforms first)
	cp dist/agent-telegram-$$(go env GOOS)-$$(go env GOARCH)* bin/agent-telegram 2>/dev/null || true
	npm pack

npm-publish: build-all ## Publish to npm (requires npm login)
	cp dist/agent-telegram-$$(go env GOOS)-$$(go env GOARCH)* bin/agent-telegram 2>/dev/null || true
	npm publish
