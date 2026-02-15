# Variables
BINARY_NAME=fiber-starter
BUILD_DIR=build
COVERAGE_DIR=coverage

.PHONY: all help build build-prod run dev test coverage lint fmt vet clean \
        migrate migrate-rollback seed seed-random routes jwt schedule \
        docs install-tools deps init

# Default target
all: help

help: ## 显示帮助信息
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# --- Development ---

dev: ## 启动开发服务器 (自动检测 air)
	@command -v air >/dev/null 2>&1 && air || go run main.go

run: ## 直接运行应用
	@go run main.go

# --- Build ---

build: ## 构建应用
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Build success: $(BUILD_DIR)/$(BINARY_NAME)"

build-prod: ## 构建生产版本 (压缩体积)
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "Production build success: $(BUILD_DIR)/$(BINARY_NAME)"

clean: ## 清理构建文件
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)

# --- Test & Quality ---

test: ## 运行测试
	@go test -v ./...

coverage: ## 生成测试覆盖率报告
	@mkdir -p $(COVERAGE_DIR)
	@go test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

lint: ## 运行代码检查 (golangci-lint)
	@golangci-lint run | grep -E "nilerr|noctx|staticcheck|gocognit|gocyclo|funlen|errcheck"

lint-strict: ## 运行代码检查 (golangci-lint)
	@golangci-lint run

lint-fix: ## 运行代码检查并自动修复
	@golangci-lint run --fix

fmt: ## 格式化代码
	@go fmt ./...

vet: ## 静态检查
	@go vet ./...

check: fmt vet lint test ## 运行所有检查

# --- Database & CLI ---

migrate: ## 运行数据库迁移
	@go run cli.go migrate run

migrate-rollback: ## 回滚数据库迁移
	@go run cli.go migrate rollback

seed: ## 运行数据库填充
	@go run cli.go seed run

seed-random: ## 生成随机测试数据 (默认 10 条)
	@go run cli.go seed run:random 10

routes: ## 显示所有路由
	@go run cli.go routes

jwt: ## 生成新的 JWT 密钥
	@go run cli.go jwt:generate

schedule: ## 运行定时任务调度器
	@go run cli.go schedule:run

cli: ## 打开命令行工具
	@go run cli.go --help

# --- Tools & Setup ---

init: install-tools deps ## 初始化项目
	@[ -f .env ] || cp .env.example .env

deps: ## 下载并整理依赖
	@go mod download && go mod tidy

install-tools: ## 安装开发工具
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

docs: ## 生成 Swagger 文档
	@swag init -g main.go -o docs
