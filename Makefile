build:
	@go build -o bin/marketplace

run: build
	@./bin/marketplace

test:
	@go test -v ./...