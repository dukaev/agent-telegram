.PHONY: lint lint-fix build run test clean

REVIVE = $(shell go env GOPATH)/bin/revive
AIR = $(shell go env GOPATH)/bin/air
GCL_CUSTOM = ./bin/custom-gcl

lint:
	golangci-lint run ./...
	@$(REVIVE) -config .revive.toml ./...
	@$(GCL_CUSTOM) run -c .golangci.custom.yaml ./...

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
