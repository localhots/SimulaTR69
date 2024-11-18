.PHONY: run build lint test

run:
	go run cmd/server/main.go

build:
	docker build -t ghcr.io/localhots/simulatr69 .

lint:
	golangci-lint run
test:
	go test -v ./...
