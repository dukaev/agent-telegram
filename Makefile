.PHONY: lint lint-fix build run test clean

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

build:
	go build -o agent-telegram .

run:
	go run .

test:
	go test -v ./...

clean:
	go clean
	rm -f agent-telegram

login-mock:
	go run main.go login -mock
