# Variables
BINARY_NAME=fiber-starter
BUILD_DIR=build
COVERAGE_DIR=coverage

.PHONY: all help build build-cli build-prod run dev test coverage lint fmt vet clean \
        migrate migrate-rollback seed seed-random routes jwt schedule \
	        docs install-tools deps init sync \
	        atlas-status atlas-history atlas-repair atlas-reset \
	        atlas-diff atlas-apply atlas-diff-postgres atlas-apply-postgres \
	        atlas-diff-sqlite atlas-apply-sqlite atlas-lint atlas-inspect

# Default target
all: help

help: ## 显示帮助信息
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# --- Development ---

dev: ## 启动开发服务器 (自动检测 air)
	@command -v air >/dev/null 2>&1 && air || go run ./cmd/server

run: ## 直接运行应用
	@go run ./cmd/server

# --- Build ---

build: ## 构建应用
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Build success: $(BUILD_DIR)/$(BINARY_NAME)"

build-cli: ## 构建 CLI 工具
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME)-cli ./cmd/cli
	@echo "Build success: $(BUILD_DIR)/$(BINARY_NAME)-cli"

build-prod: ## 构建生产版本 (压缩体积)
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
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
	@golangci-lint run

lint-quick: ## 运行代码检查 (只显示常见问题)
	@golangci-lint run | grep -E "nilerr|noctx|staticcheck|gocognit|gocyclo|funlen|errcheck" || true

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
	@go run ./cmd/cli migrate run

migrate-rollback: ## 回滚数据库迁移
	@go run ./cmd/cli migrate rollback

seed: ## 运行数据库填充
	@go run ./cmd/cli seed run

seed-random: ## 生成随机测试数据 (默认 10 条)
	@go run ./cmd/cli seed run:random 10

routes: ## 显示所有路由
	@go run ./cmd/cli routes

jwt: ## 生成新的 JWT 密钥
	@go run ./cmd/cli jwt:generate

schedule: ## 运行定时任务调度器
	@go run ./cmd/cli schedule:run

cli: ## 打开命令行工具
	@go run ./cmd/cli --help

# --- Tools & Setup ---

init: install-tools deps ## 初始化项目
	@[ -f .env ] || cp .env.example .env

sync: ## 同步并整理依赖
	@echo "Cleaning and syncing dependencies..."
	go mod tidy
	go mod vendor
	@echo "Done. Please check GoLand settings for Vendoring."

deps: ## 下载并整理依赖
	@go mod download && go mod tidy

install-tools: ## 安装开发工具
	@go install github.com/air-verse/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

atlas-diff-postgres: ## 生成 PostgreSQL 迁移（NAME=xxx）
	@atlas migrate diff $(NAME) --env postgres

atlas-apply-postgres: ## 应用 PostgreSQL 迁移（依赖 DATABASE_URL）
	@atlas migrate apply --env postgres

atlas-diff-sqlite: ## 生成 SQLite 迁移（NAME=xxx）
	@atlas migrate diff $(NAME) --env sqlite

atlas-apply-sqlite: ## 应用 SQLite 迁移
	@atlas migrate apply --env sqlite

atlas-status: ## 显示迁移状态（默认 postgres，ENV=sqlite 可切换）
	@atlas migrate status --env $(or $(ENV),postgres)

atlas-history: ## 显示迁移历史（默认 postgres，ENV=sqlite 可切换）
	@atlas migrate history --env $(or $(ENV),postgres)

atlas-repair: ## 修复迁移表（默认 postgres，ENV=sqlite 可切换）
	@atlas migrate repair --env $(or $(ENV),postgres)

atlas-reset: ## 重置数据库并重新应用所有迁移（默认 postgres，ENV=sqlite 可切换）
	@atlas migrate reset --env $(or $(ENV),postgres)

atlas-diff: ## 生成迁移（默认 postgres，NAME=xxx，ENV=sqlite 可切换）
	@atlas migrate diff $(NAME) --env $(or $(ENV),postgres)

atlas-apply: ## 应用迁移（默认 postgres，ENV=sqlite 可切换）
	@atlas migrate apply --env $(or $(ENV),postgres)

atlas-lint: ## 检查数据库 schema（默认 postgres，ENV=sqlite 可切换）
	@atlas schema lint --env $(or $(ENV),postgres)

atlas-inspect: ## 检查当前数据库 schema（默认 postgres，ENV=sqlite 可切换）
	@atlas schema inspect --env $(or $(ENV),postgres)

docs: ## 生成 OpenAPI/Swagger 规范（由 Scalar 展示）
	@swag init -g ./cmd/server/main.go -o docs
	@rm -f docs/docs.go
