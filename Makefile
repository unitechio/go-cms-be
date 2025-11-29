.PHONY: help run build test clean migrate-up migrate-down swagger docker-up docker-down

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

run: ## Run the application
	go run cmd/server/main.go

build: ## Build the application
	go build -o bin/server cmd/server/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	rm -rf bin/ dist/ coverage.out

install-deps: ## Install dependencies
	go mod download
	go mod tidy

swagger: ## Generate swagger documentation
	swag init -g cmd/server/main.go -o docs

migrate-up: ## Run database migrations up
	go run cmd/migrate/main.go up

migrate-down: ## Run database migrations down
	go run cmd/migrate/main.go down

migrate-create: ## Create a new migration file (usage: make migrate-create name=create_users_table)
	migrate create -ext sql -dir migrations -seq $(name)

docker-up: ## Start docker containers
	docker-compose up -d

docker-down: ## Stop docker containers
	docker-compose down

docker-logs: ## View docker logs
	docker-compose logs -f

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

install-tools: ## Install development tools
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev: docker-up run ## Start development environment
