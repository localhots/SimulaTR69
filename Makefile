run:
	go run cmd/server/main.go

lint:
	golangci-lint run
test:
	go test -v ./...
