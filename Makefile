.PHONY: lint lint-fix build run test test-contracts validate-fixtures clean release-local release release-minor release-major npm-pack npm-publish setup-hooks

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

LAST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
MAJOR := $(shell echo $(LAST_TAG) | sed 's/v//' | cut -d. -f1)
MINOR := $(shell echo $(LAST_TAG) | sed 's/v//' | cut -d. -f2)
PATCH := $(shell echo $(LAST_TAG) | sed 's/v//' | cut -d. -f3)
NEXT_PATCH := v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1)))
NEXT_MINOR := v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0
NEXT_MAJOR := v$(shell echo $$(($(MAJOR)+1))).0.0

release: ## Patch release (v0.4.1 → v0.4.2)
	@echo "$(LAST_TAG) → $(NEXT_PATCH)"
	git tag $(NEXT_PATCH)
	git push origin main --tags

release-minor: ## Minor release (v0.4.1 → v0.5.0)
	@echo "$(LAST_TAG) → $(NEXT_MINOR)"
	git tag $(NEXT_MINOR)
	git push origin main --tags

release-major: ## Major release (v0.4.1 → v1.0.0)
	@echo "$(LAST_TAG) → $(NEXT_MAJOR)"
	git tag $(NEXT_MAJOR)
	git push origin main --tags

release-local: ## Build release locally (for testing)
	goreleaser release --snapshot --clean

npm-pack: ## Pack npm package
	npm pack

npm-publish: ## Publish to npm (requires npm login)
	npm publish

setup-hooks: ## Configure git to use .githooks/ for hooks
	git config core.hooksPath .githooks
