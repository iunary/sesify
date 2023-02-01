.PHONY: wire clean fromat run critic security lint build 
APP_NAME = sesify
BUILD_DIR = $(PWD)/build
IMAGE = simply
TAG = latest

clean:
	@rm -rf $(BUILD_DIR)
	
format:
	@go fmt ./...

critic:
	@gocritic check -enableAll ./...

security:
	@gosec ./...

lint:
	@golangci-lint run ./...

run: 
	@go run ./...

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) .