build:
	@go build -o bin/marketplace

run: build
	@./bin/marketplace

database:
	@docker-compose up -d

test:
	@go test -v ./...