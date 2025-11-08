.PHONY: help build run test clean docker-up docker-down migrate seed swagger

help:
	@echo "Available targets:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application locally"
	@echo "  test            - Run tests"
	@echo "  test-cover      - Run tests with coverage"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-up       - Start Docker services"
	@echo "  docker-down     - Stop Docker services"
	@echo "  migrate-up      - Run database migrations up"
	@echo "  migrate-down    - Rollback last migration"
	@echo "  migrate-version - Check current migration version"
	@echo "  migrate-create  - Create new migration file"
	@echo "  seed            - Seed database with test data"
	@echo "  swagger         - Generate Swagger documentation"
	@echo "  lint            - Run linters"

build:
	@echo "Building application..."
	go build -o bin/api cmd/api/main.go

run:
	@echo "Running application..."
	go run cmd/api/main.go

test:
	@echo "Running tests..."
	go test -v -race ./...

test-cover:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	docker-compose logs -f

migrate-up:
	@echo "Running migrations up..."
	go run cmd/migrate/main.go -cmd up -path migrations

migrate-down:
	@echo "Running migrations down..."
	go run cmd/migrate/main.go -cmd down -steps 1 -path migrations

migrate-version:
	@echo "Checking migration version..."
	go run cmd/migrate/main.go -cmd version -path migrations

migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

seed:
	@echo "Seeding database..."
	go run cmd/seed/main.go

swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/api/main.go -o docs

lint:
	@echo "Running linters..."
	golangci-lint run

.DEFAULT_GOAL := help
