# Makefile for Assets Discovery System

.PHONY: build clean install deps test lint fmt run-live run-offline config prepare package help

# 变量定义
BINARY_NAME=assets_discovery
BUILD_DIR=./build
GO_FILES=$(shell find . -name "*.go" -type f)

# 默认目标
all: build

# 编译项目
build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./main.go
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# 安装依赖
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# 清理构建文件
clean:
	@echo "Cleaning build files..."
	@rm -rf $(BUILD_DIR)
	@rm -rf output
	@rm -rf logs
	@rm -rf release
	@go clean

# 安装到系统
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin/"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# 运行测试
test:
	@echo "Running tests..."
	@go test -v ./...

# 检查代码质量
lint:
	@echo "Running linter..."
	@go vet ./...
	@gofmt -l .

# 格式化代码
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# 实时监听模式（需要root权限）
run-live: build
	@echo "Starting live capture mode..."
	@echo "Usage: sudo $(BUILD_DIR)/$(BINARY_NAME) live -i <interface>"
	@echo "Example: sudo $(BUILD_DIR)/$(BINARY_NAME) live -i eth0"

# 离线分析模式
run-offline: build
	@echo "Starting offline analysis mode..."
	@echo "Usage: $(BUILD_DIR)/$(BINARY_NAME) offline -f <pcap_file>"
	@echo "Example: $(BUILD_DIR)/$(BINARY_NAME) offline -f capture.pcap"

# 创建配置文件
config:
	@echo "Creating configuration file..."
	@cp config.yaml assets_discovery.yaml
	@echo "Configuration created: assets_discovery.yaml"

# 安装系统依赖（Ubuntu/Debian）
install-deps-ubuntu:
	@echo "Installing system dependencies for Ubuntu/Debian..."
	@sudo apt-get update
	@sudo apt-get install -y libpcap-dev build-essential

# 安装系统依赖（CentOS/RHEL）
install-deps-centos:
	@echo "Installing system dependencies for CentOS/RHEL..."
	@sudo yum install -y libpcap-devel gcc make

# 创建输出目录
prepare:
	@echo "Preparing directories..."
	@mkdir -p ./output
	@mkdir -p ./logs

# 打包发布
package: build
	@echo "Creating release package..."
	@mkdir -p release
	@cp $(BUILD_DIR)/$(BINARY_NAME) release/
	@cp config.yaml release/
	@cp README.md release/
	@cd release && tar -czf $(BINARY_NAME)-$(shell date +%Y%m%d).tar.gz *
	@echo "Package created: release/$(BINARY_NAME)-$(shell date +%Y%m%d).tar.gz"

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  deps               - Install Go dependencies"
	@echo "  clean              - Clean build files"
	@echo "  install            - Install to system (/usr/local/bin)"
	@echo "  test               - Run tests"
	@echo "  lint               - Run code linter"
	@echo "  fmt                - Format code"
	@echo "  run-live           - Show usage for live capture mode"
	@echo "  run-offline        - Show usage for offline analysis mode"
	@echo "  config             - Create example configuration"
	@echo "  install-deps-ubuntu - Install system deps on Ubuntu/Debian"
	@echo "  install-deps-centos - Install system deps on CentOS/RHEL"
	@echo "  prepare            - Create necessary directories"
	@echo "  package            - Create release package"
	@echo "  help               - Show this help message"
