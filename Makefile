# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install # Added for swag tool

# Docker parameters
DOCKER_IMAGE=messaging-service
DOCKER_TAG=latest

# Docker Compose
DOCKER_COMPOSE=docker-compose

# Binary name
BINARY_NAME=messaging-service

# Phony targets
.PHONY: setup run test clean help swagger docs docker-build docker-run docker-stop docker-clean docker-prod docker-prod-stop docker-prod-logs docker-dev

help:
	@echo "Available commands:"
	@echo "  setup    - Initialize project dependencies"
	@echo "  run      - Start the messaging service"
	@echo "  test     - Run all tests"
	@echo "  clean    - Clean up build artifacts"
	@echo "  swagger  - Generate Swagger documentation"
	@echo "  docs     - Generate Swagger documentation"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  docker-stop  - Stop Docker container"
	@echo "  docker-clean - Clean Docker resources"
	@echo "  docker-up    - Start the full stack with Docker Compose"
	@echo "  docker-down  - Stop the full stack"
	@echo "  docker-logs  - View logs"
	@echo "  help     - Show this help message"

setup:
	@echo "Setting up project..."
	@go mod tidy
	@echo "Project setup complete!"

run:
	@echo "Starting messaging service..."
	@go run cmd/server/main.go

test:
	@echo "Running tests..."
	@echo "Starting test database if not running..."
	@docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running unit tests..."
	@go test ./internal/... -v
	@echo "Running integration tests..."
	@go test ./tests/... -v
	@echo "Running test script..."
	@./bin/test.sh

clean:
	@echo "Cleaning up..."
	@go clean
	@rm -f $(BINARY_NAME)

swagger:
	@echo "Generating Swagger documentation..."
	@$(GOINSTALL) github.com/swaggo/swag/cmd/swag@latest
	@swag init -g cmd/server/main.go
	@echo "Swagger documentation generated in docs/ directory"

docs: swagger
	@echo "Swagger documentation generated!"
	@echo "Start the application with 'make run' and visit:"
	@echo "  http://localhost:8080/swagger/index.html"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully!"

docker-run:
	@echo "Running Docker container..."
	@docker run -d --name messaging-service-container \
		-p 8080:8080 \
		-e DATABASE_HOST=host.docker.internal \
		-e DATABASE_PORT=5432 \
		-e DATABASE_NAME=messaging_service \
		-e DATABASE_USER=messaging_user \
		-e DATABASE_PASSWORD=messaging_password \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop:
	@echo "Stopping Docker container..."
	@docker stop messaging-service-container || true
	@docker rm messaging-service-container || true

docker-clean:
	@echo "Cleaning Docker resources..."
	@docker stop messaging-service-container || true
	@docker rm messaging-service-container || true
	@docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) || true
	@docker system prune -f

docker-up:
	@echo "Starting the full stack..."
	@docker-compose up -d
	@echo "Stack started!"
	@echo "API available at: http://localhost:8080"
	@echo "Swagger docs at: http://localhost:8080/swagger/index.html"
	@echo "Database available at: localhost:5432"

docker-down:
	@echo "Stopping the full stack..."
	@docker-compose down
	@echo "Stack stopped!"

docker-logs:
	@echo "Showing logs..."
	@docker-compose logs -f