# Variables
-include .buildconfig
export PATH := /opt/homebrew/bin:/usr/local/bin:$(PATH)
GO ?= go
GOFMT ?= gofmt
GOLANGCI_LINT ?= golangci-lint
SWAG ?= swag
ATLAS ?= atlas
BUILD_DIR ?= build
COVERAGE_DIR ?= coverage
LINT_CACHE_HOME ?= /tmp/fiber-starter-cache
LINT_GOCACHE ?= /tmp/fiber-starter-gocache
SERVER_BINARY_NAME ?= fiber-starter
CLI_BINARY_NAME ?= fiber-starter-cli
APP_LOG_DIR ?= storage/logs
DEPLOY_DIR ?= deploy
SERVER_MAIN ?= ./cmd/server
CLI_MAIN ?= ./cmd/cli
SWAG_MAIN ?= ./cmd/server/main.go
SERVER_RUN = $(GO) run $(SERVER_MAIN)
CLI_RUN = $(GO) run $(CLI_MAIN)

.DEFAULT_GOAL := help

define run_golangci_lint
	@LINT_BIN="$$(command -v $(GOLANGCI_LINT) 2>/dev/null || true)"; \
	if [ -z "$$LINT_BIN" ] && [ -x /opt/homebrew/bin/$(GOLANGCI_LINT) ]; then LINT_BIN=/opt/homebrew/bin/$(GOLANGCI_LINT); fi; \
	if [ -z "$$LINT_BIN" ] && [ -x /usr/local/bin/$(GOLANGCI_LINT) ]; then LINT_BIN=/usr/local/bin/$(GOLANGCI_LINT); fi; \
	if [ -z "$$LINT_BIN" ]; then echo "$(GOLANGCI_LINT) is not installed"; exit 1; fi; \
	mkdir -p $(LINT_GOCACHE) $(LINT_CACHE_HOME); \
	HOME=/tmp XDG_CACHE_HOME=$(LINT_CACHE_HOME) GOCACHE=$(LINT_GOCACHE) GOFLAGS=-mod=vendor "$$LINT_BIN" $(1)
endef

define build_binary
	@GOFLAGS=-mod=mod $(1) $(GO) build $(2) -o $(BUILD_DIR)/$(3) $(4)
	@echo "$(5): $(BUILD_DIR)/$(3)"
endef

.PHONY: all help build build-cli build-prod build-dir coverage-dir config run dev test coverage lint lint-strict fmt vet clean \
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
	@command -v air >/dev/null 2>&1 && air || $(SERVER_RUN)

run: ## 直接运行应用
	@$(SERVER_RUN)

# --- Build ---

build-dir:
	@mkdir -p $(BUILD_DIR)

coverage-dir:
	@mkdir -p $(COVERAGE_DIR)

build: build-dir ## 构建应用
	$(call build_binary,,,$(SERVER_BINARY_NAME),$(SERVER_MAIN),Build success)

build-cli: build-dir ## 构建 CLI 工具
	$(call build_binary,,,$(CLI_BINARY_NAME),$(CLI_MAIN),Build success)

build-prod: build-dir ## 构建生产版本 (压缩体积)
	$(call build_binary,CGO_ENABLED=0 GOOS=linux GOARCH=amd64,-ldflags="-w -s",$(SERVER_BINARY_NAME),$(SERVER_MAIN),Production build success)

config: ## 显示当前构建配置
	@printf "BUILD_DIR=%s\nCOVERAGE_DIR=%s\nSERVER_BINARY_NAME=%s\nCLI_BINARY_NAME=%s\nAPP_LOG_DIR=%s\nDEPLOY_DIR=%s\n" \
		"$(BUILD_DIR)" "$(COVERAGE_DIR)" "$(SERVER_BINARY_NAME)" "$(CLI_BINARY_NAME)" "$(APP_LOG_DIR)" "$(DEPLOY_DIR)"

clean: ## 清理构建文件
	@rm -rf $(BUILD_DIR) $(COVERAGE_DIR)

# --- Test & Quality ---

test: ## 运行测试
	@$(GO) test -v ./...

coverage: coverage-dir ## 生成测试覆盖率报告
	@$(GO) test -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report: $(COVERAGE_DIR)/coverage.html"

lint: ## 运行代码检查 (golangci-lint)
	$(call run_golangci_lint,run)

lint-quick: ## 运行代码检查 (只显示常见问题)
	$(call run_golangci_lint,run | grep -E "nilerr|noctx|staticcheck|gocognit|gocyclo|funlen|errcheck" || true)

lint-strict: lint ## 运行代码检查 (golangci-lint)

lint-fix: ## 运行代码检查并自动修复
	$(call run_golangci_lint,run --fix)

fmt: ## 格式化代码
	@files="$$(find . -name '*.go' -not -path './vendor/*' -not -path './build/*' -not -path './coverage/*')"; \
	if [ -n "$$files" ]; then $(GOFMT) -w $$files; fi

vet: ## 静态检查
	@$(GO) vet ./...

check: fmt vet lint test ## 运行所有检查

# --- Database & CLI ---

migrate: ## 运行数据库迁移
	@$(CLI_RUN) migrate run

migrate-rollback: ## 回滚数据库迁移
	@$(CLI_RUN) migrate rollback

seed: ## 运行数据库填充
	@$(CLI_RUN) seed run

seed-random: ## 生成随机测试数据 (默认 10 条)
	@$(CLI_RUN) seed run:random 10

routes: ## 显示所有路由
	@$(CLI_RUN) routes

jwt: ## 生成新的 JWT 密钥
	@$(CLI_RUN) jwt:generate

schedule: ## 运行定时任务调度器
	@$(CLI_RUN) schedule:run

cli: ## 打开命令行工具
	@$(CLI_RUN) --help

# --- Tools & Setup ---

init: install-tools deps ## 初始化项目
	@[ -f .env ] || cp .env.example .env

sync: ## 同步并整理依赖
	@echo "Cleaning and syncing dependencies..."
	go mod tidy
	go mod vendor
	@echo "Done. Please check GoLand settings for Vendoring."

deps: ## 下载并整理依赖
	@$(GO) mod download && $(GO) mod tidy

install-tools: ## 安装开发工具
	@$(GO) install github.com/air-verse/air@latest
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest

atlas-diff-postgres: ## 生成 PostgreSQL 迁移（NAME=xxx）
	@$(ATLAS) migrate diff $(NAME) --env postgres

atlas-apply-postgres: ## 应用 PostgreSQL 迁移（依赖 DATABASE_URL）
	@$(ATLAS) migrate apply --env postgres

atlas-diff-sqlite: ## 生成 SQLite 迁移（NAME=xxx）
	@$(ATLAS) migrate diff $(NAME) --env sqlite

atlas-apply-sqlite: ## 应用 SQLite 迁移
	@$(ATLAS) migrate apply --env sqlite

atlas-status: ## 显示迁移状态（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) migrate status --env $(or $(ENV),postgres)

atlas-history: ## 显示迁移历史（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) migrate history --env $(or $(ENV),postgres)

atlas-repair: ## 修复迁移表（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) migrate repair --env $(or $(ENV),postgres)

atlas-reset: ## 重置数据库并重新应用所有迁移（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) migrate reset --env $(or $(ENV),postgres)

atlas-diff: ## 生成迁移（默认 postgres，NAME=xxx，ENV=sqlite 可切换）
	@$(ATLAS) migrate diff $(NAME) --env $(or $(ENV),postgres)

atlas-apply: ## 应用迁移（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) migrate apply --env $(or $(ENV),postgres)

atlas-lint: ## 检查数据库 schema（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) schema lint --env $(or $(ENV),postgres)

atlas-inspect: ## 检查当前数据库 schema（默认 postgres，ENV=sqlite 可切换）
	@$(ATLAS) schema inspect --env $(or $(ENV),postgres)

docs: ## 生成 OpenAPI/Swagger 规范（由 Scalar 展示）
	@$(SWAG) init -g $(SWAG_MAIN) -o docs
	@rm -f docs/docs.go
