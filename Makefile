build:
	@go build -o bin/ionic ./cmd/ionic-server

run:
	@go run ./cmd/ionic-server/main.go

test:
	@go test -v ./...

format:
	@gofmt -s -w ./...