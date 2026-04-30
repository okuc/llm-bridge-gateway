# LLM Bridge Gateway Makefile

# 变量定义
APP_NAME := llm-bridge-gateway
VERSION := 1.0.0
BUILD_DIR := ./build
CMD_DIR := ./cmd/gateway

# Go 命令
GO := go
GOBUILD := $(GO) build
GOTEST := $(GO) test
GOVET := $(GO) vet
GOMOD := $(GO) mod

# 默认目标
.PHONY: all
all: clean build

# 构建
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# 构建Linux版本
.PHONY: build-linux
build-linux:
	@echo "Building $(APP_NAME) for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)/main.go
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)-linux-amd64"

# 构建所有平台
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_DIR)/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_DIR)/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_DIR)/main.go
	@echo "Build complete"

# 运行
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	$(BUILD_DIR)/$(APP_NAME) -config ./config/config.yaml

# 运行（开发模式）
.PHONY: dev
dev:
	@echo "Running in development mode..."
	$(GO) run $(CMD_DIR)/main.go -config ./config/config.yaml

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 测试覆盖率
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# 依赖整理
.PHONY: tidy
tidy:
	@echo "Running go mod tidy..."
	$(GOMOD) tidy

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build"
	@echo "  build        - Build the application"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-all    - Build for all platforms"
	@echo "  run          - Build and run"
	@echo "  dev          - Run in development mode"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  vet          - Run go vet"
	@echo "  tidy         - Run go mod tidy"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"
