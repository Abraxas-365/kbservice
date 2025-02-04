# Environment variables
export OPENAI_API_KEY ?= your-default-openai-key-here
export DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/vectordb

.PHONY: up down clean run psql init-and-run logs

# Start the database
up:
	docker-compose up -d

# Stop the database
down:
	docker-compose down

# Remove all data
clean:
	docker-compose down -v

# Connect to the database
psql:
	docker exec -it kbservice-postgres-1 psql -U postgres -d vectordb

# Run the example
run:
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY is not set"; \
		exit 1; \
	fi
	go run examples/basic/main.go

# Initialize the database and run the example
init-and-run: up
	@echo "Waiting for database to be ready..."
	@sleep 5
	@if [ -z "$(OPENAI_API_KEY)" ]; then \
		echo "Error: OPENAI_API_KEY is not set"; \
		exit 1; \
	fi
	go run examples/basic/main.go

# View database logs
logs:
	docker-compose logs -f postgres
