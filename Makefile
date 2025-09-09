APP_NAME := hermes
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP_NAME)

.PHONY: all build test lint run clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) .

test:
	go test ./... -race -cover

lint:
	go vet ./...
	go fmt ./...

run:
	go run . start-tui

clean:
	rm -rf $(BIN_DIR)

