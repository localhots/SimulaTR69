.PHONY: run build lint test

run:
	go run cmd/server/main.go

build:
	docker build -t ghcr.io/localhots/simulatr69:latest .

lint:
	golangci-lint run

test:
	go test -v ./...

bench-noise:
	go test -bench=. ./datamodel/noise

gen-noise-preview:
	go test -tags=preview -v ./datamodel/noise -run TestGeneratePreviews
