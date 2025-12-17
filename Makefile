# Go 参数
GOCMD=go
BINARY_NAME=fiber-starter
BUILD_DIR=build
COVERAGE_DIR=coverage

.PHONY: help dev run build test clean deps lint fmt vet

# 默认目标
help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 开发命令
dev: ## 启动开发服务器（带热重载）
	@echo "启动开发服务器..."
	@if command -v air >/dev/null 2>&1; then air; else go run main.go; fi

run: ## 运行应用程序
	@echo "运行应用程序..."
	$(GOCMD) run main.go

# 构建命令
build: ## 构建应用程序
	@echo "构建应用程序..."
	@mkdir -p $(BUILD_DIR)
	$(GOCMD) build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

build-prod: ## 构建生产版本
	@echo "构建生产版本..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOCMD) build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# 测试命令
test: ## 运行测试
	@echo "运行测试..."
	$(GOCMD) test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试并生成覆盖率报告..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOCMD) test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html

# 代码质量
fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOCMD) fmt ./...
	@if [ -f "$$GOPATH/bin/goimports" ]; then $$GOPATH/bin/goimports -w .; elif [ -f "$$HOME/go/bin/goimports" ]; then $$HOME/go/bin/goimports -w .; fi

vet: ## 代码静态检查
	@echo "运行代码静态检查..."
	$(GOCMD) vet ./...

lint: ## 运行代码检查工具
	@echo "运行代码检查工具..."
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run --disable-all -E gofmt,govet,errcheck,staticcheck,unused --fast; else echo "golangci-lint 未安装，使用 make install-tools 安装"; fi

lint-fix: ## 运行代码检查工具并自动修复
	@echo "运行代码检查工具并自动修复..."
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run --disable-all -E gofmt,govet,errcheck,staticcheck,unused --fix --fast; else echo "golangci-lint 未安装，使用 make install-tools 安装"; fi

format: fmt lint-fix ## 格式化代码并修复问题 (类似 Laravel Pint)
	@echo "✅ 代码格式化完成"

check: fmt vet test ## 运行所有代码检查

# 依赖管理
deps: ## 下载依赖
	@echo "下载依赖..."
	$(GOCMD) mod download
	$(GOCMD) mod tidy

deps-update: ## 更新依赖
	@echo "更新依赖..."
	$(GOCMD) get -u ./...
	$(GOCMD) mod tidy

# 工具安装
install-tools: ## 安装开发工具
	@echo "安装开发工具..."
	$(GOCMD) install github.com/air-verse/air@latest
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest

# 清理命令
clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f $(BINARY_NAME)

# 数据库命令
db-migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	@if [ -f "database/migrations/migrate.go" ]; then $(GOCMD) run database/migrations/migrate.go up; else echo "迁移文件不存在"; fi

db-seed: ## 运行数据库种子
	@echo "运行数据库种子..."
	@if [ -f "database/seeders/seed.go" ]; then $(GOCMD) run database/seeders/seed.go; else echo "种子文件不存在"; fi

# 生成命令
generate: ## 生成代码
	@echo "生成代码..."
	$(GOCMD) generate ./...

jwt: ## 生成新的 JWT 密钥
	@echo "生成新的 JWT 密钥..."
	$(GOCMD) run cli.go jwt:generate

routes: ## 显示所有路由
	@$(GOCMD) run cli.go routes

cli: ## 打开命令行工具
	@echo "💻 命令行工具"
	$(GOCMD) run cli.go --help

docs: ## 生成 API 文档
	@echo "生成 API 文档..."
	@if command -v swag >/dev/null 2>&1; then swag init -g main.go -o docs; else echo "swag 未安装"; fi

# 版本信息
version: ## 显示版本信息
	@echo "版本信息:"
	@$(GOCMD) version
	@echo "构建时间: $(shell date)"
	@echo "Git 提交: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"

# 初始化项目
init: ## 初始化新项目
	@echo "初始化项目..."
	$(MAKE) install-tools
	$(MAKE) deps
	@if [ ! -f ".env" ]; then cp .env.example .env; echo "已创建 .env 文件"; fi

# 快速命令
quick-test: fmt vet ## 快速检查（格式化和静态检查）
	@echo "快速检查完成"

ci: deps check test-coverage ## CI 流水线
	@echo "CI 流水线执行完成"