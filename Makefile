.PHONY: build run clean test fmt lint

APP_NAME := mailfake
BUILD_DIR := bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd

run:
	go run ./cmd

clean:
	rm -rf $(BUILD_DIR)

fmt:
	go fmt ./...

lint:
	golangci-lint run

test:
	go test -v ./...

# Сборка для разных платформ
build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux ./cmd

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin ./cmd

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME).exe ./cmd

build-all: build-linux build-mac build-windows

# Docker
docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker run -p 1025:1025 -p 8025:8025 $(APP_NAME)
