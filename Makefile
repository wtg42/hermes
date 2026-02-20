APP_NAME := hermes
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

.PHONY: all build test lint run clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

test:
	@echo "Checking Docker availability..."
	@docker ps > /dev/null 2>&1 || (echo "Error: Docker is not available. Please install Docker to run tests with Mailpit."; exit 1)
	@echo "Starting Mailpit service..."
	@docker-compose down > /dev/null 2>&1; docker-compose up -d > /dev/null 2>&1
	@echo "Waiting for Mailpit to be ready..."
	@sleep 2
	@echo "Running tests (including integration tests)..."
	@go test ./... -race -cover -tags integration; TEST_EXIT=$$?; \
	echo "Cleaning up Mailpit..."; \
	docker-compose down > /dev/null 2>&1; \
	exit $$TEST_EXIT

lint:
	go vet ./...
	go fmt ./...

run:
	go run . start-tui

clean:
	rm -rf $(BIN_DIR)

