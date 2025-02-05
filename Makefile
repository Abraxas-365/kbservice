# Environment variables
export OPENAI_API_KEY ?= your-default-openai-key-here
export DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/vectordb

.PHONY: up down clean psql run-llm run-kb init-db logs help test build

# Show help
help:
	@echo "Available commands:"
	@echo "  make up              - Start the database container"
	@echo "  make down            - Stop the database container"
	@echo "  make clean           - Remove all data and containers"
	@echo "  make psql            - Connect to PostgreSQL database"
	@echo "  make run-llm         - Run LLM example"
	@echo "  make run-kb          - Run Knowledge Base example"
	@echo "  make init-db         - Initialize database"
	@echo "  make logs            - View database logs"
	@echo "  make test            - Run tests"
	@echo "  make build           - Build the project"

# Database management
up:
	@echo "Starting database container..."
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5

down:
	@echo "Stopping database container..."
	docker-compose down

clean:
	@echo "Removing all containers and volumes..."
	docker-compose down -v

psql:
	@echo "Connecting to PostgreSQL..."
	docker exec -it kbservice-postgres-1 psql -U postgres -d vectordb

logs:
	@echo "Showing database logs..."
	docker-compose logs -f postgres

# Check OpenAI API key
check-openai-key:
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY is not set"; \
		exit 1; \
	fi

# Initialize database
init-db: up
	@echo "Initializing database..."
	@sleep 5

# Run examples
run-llm: check-openai-key
	@echo "Running LLM example..."
	go run examples/llm/main.go

run-kb: check-openai-key up
	@echo "Running Knowledge Base example..."
	go run examples/kb/main.go

# Development
test:
	@echo "Running tests..."
	go test ./... -v

build:
	@echo "Building project..."
	go build -v ./...

# Combined commands
init-and-run-kb: init-db run-kb

# Default target
.DEFAULT_GOAL := help
