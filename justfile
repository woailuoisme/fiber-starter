set dotenv-load
set dotenv-filename := ".buildconfig"
set shell := ["bash", "-euo", "pipefail", "-c"]

export PATH := "/opt/homebrew/bin:/usr/local/bin:" + env("PATH", "")

ROOT := justfile_directory()

export GOFLAGS := env("GOFLAGS", "-mod=mod")

GO := env("GO", "go")
GOFUMPT := env("GOFUMPT", "gofumpt")
GOLANGCI_LINT := env("GOLANGCI_LINT", "golangci-lint")
K6 := env("K6", "k6")
ATLAS := env("ATLAS", "atlas")
ATLAS_MIGRATE := ATLAS + " migrate"
ATLAS_SCHEMA := ATLAS + " schema"
BUILD_DIR := env("BUILD_DIR", "build")
COVERAGE_DIR := env("COVERAGE_DIR", "tests/coverage")
LINT_CACHE_HOME := env("LINT_CACHE_HOME", ROOT / ".cache/golangci-lint")
LINT_GOCACHE := env("LINT_GOCACHE", ROOT / ".cache/go-build")
SERVER_BINARY_NAME := env("SERVER_BINARY_NAME", "lfiber")
CLI_BINARY_NAME := env("CLI_BINARY_NAME", "lfiber-cli")
APP_LOG_DIR := env("APP_LOG_DIR", "storage/logs")
DEPLOY_DIR := env("DEPLOY_DIR", "deploy")
APP_MAIN := env("APP_MAIN", "./cmd/app")
APP_RUN := GO + " run " + APP_MAIN
ENV := env("ENV", "postgres")
NAME := env("NAME", "")
CMD := env("CMD", "")

LINT_RUN := "XDG_CACHE_HOME=" + LINT_CACHE_HOME + " GOCACHE=" + LINT_GOCACHE + " " + GOLANGCI_LINT + " run"

# 显示帮助信息
default:
    @just --list

# 显示帮助信息
help: default

# 启动开发服务器 (自动检测 air)
dev:
    @command -v air >/dev/null 2>&1 && air || {{ APP_RUN }} serve

# 直接运行应用
run:
    @{{ APP_RUN }} serve

_build-dir:
    @mkdir -p {{ BUILD_DIR }}

_coverage-dir:
    @mkdir -p {{ COVERAGE_DIR }}

# 构建应用
build: _build-dir
    @{{ GO }} build -o {{ BUILD_DIR }}/{{ SERVER_BINARY_NAME }} {{ APP_MAIN }}
    @echo "Build success: {{ BUILD_DIR }}/{{ SERVER_BINARY_NAME }}"

# 构建 CLI 工具（同一二进制）
build-cli: _build-dir
    @{{ GO }} build -o {{ BUILD_DIR }}/{{ CLI_BINARY_NAME }} {{ APP_MAIN }}
    @echo "Build success: {{ BUILD_DIR }}/{{ CLI_BINARY_NAME }}"

# 构建生产版本 (压缩体积)
build-prod: _build-dir
    @CGO_ENABLED=0 GOOS=linux GOARCH=amd64 {{ GO }} build -ldflags="-w -s" -o {{ BUILD_DIR }}/{{ SERVER_BINARY_NAME }} {{ APP_MAIN }}
    @echo "Production build success: {{ BUILD_DIR }}/{{ SERVER_BINARY_NAME }}"

# 显示当前构建配置
config:
    @printf "BUILD_DIR=%s\nCOVERAGE_DIR=%s\nSERVER_BINARY_NAME=%s\nCLI_BINARY_NAME=%s\nAPP_LOG_DIR=%s\nDEPLOY_DIR=%s\n" \
        "{{ BUILD_DIR }}" "{{ COVERAGE_DIR }}" "{{ SERVER_BINARY_NAME }}" "{{ CLI_BINARY_NAME }}" "{{ APP_LOG_DIR }}" "{{ DEPLOY_DIR }}"

# 清理构建文件
clean:
    @rm -rf {{ BUILD_DIR }} {{ COVERAGE_DIR }}

# 运行测试
test:
    @{{ GO }} test -v ./...

# 生成测试覆盖率报告
coverage:
    @sh scripts/coverage.sh

_lint-setup:
    @mkdir -p {{ LINT_GOCACHE }} {{ LINT_CACHE_HOME }}

# 运行代码检查 (golangci-lint)
lint: _lint-setup
    @{{ LINT_RUN }}

# 运行代码检查 (只显示常见问题)
lint-quick: _lint-setup
    @{{ LINT_RUN }} | grep -E "nilerr|noctx|staticcheck|gocognit|gocyclo|funlen|errcheck" || true

# 运行代码检查 (golangci-lint)
lint-strict: lint

# 运行代码检查并自动修复
lint-fix: _lint-setup
    @{{ LINT_RUN }} --fix

# 格式化代码
fmt:
    @{{ GO }} fmt ./...

# 使用 gofumpt 格式化 Go 代码
fmt-gofumpt:
    @{{ GOFUMPT }} -w .

# 格式化 Swagger 注释
fmt-swag:
    @swag fmt

# 静态检查
vet:
    @{{ GO }} vet ./...

# 自动化格式化和静态修复
check: fmt-gofumpt lint-fix

# 运行所有检查
check-all: check test

# 运行 CLI 命令（示例：just artisan jwt:generate 或 CMD="jwt:generate" just artisan）
artisan cmd=CMD:
    @{{ APP_RUN }} {{ cmd }}

# 运行数据库迁移
migrate:
    @{{ APP_RUN }} db:migrate

# 回滚数据库迁移
migrate-rollback:
    @{{ APP_RUN }} db:rollback

# 运行数据库填充
seed:
    @{{ APP_RUN }} db:seed

# 生成随机测试数据 (默认 10 条)
seed-random:
    @{{ APP_RUN }} db:seed-random 10

# 显示所有路由
routes:
    @{{ APP_RUN }} route:list

# 生成新的 JWT 密钥
jwt:
    @{{ APP_RUN }} jwt:generate

# 运行定时任务调度器
schedule:
    @{{ APP_RUN }} schedule:run

# 打开命令行工具
cli:
    @{{ APP_RUN }} --help

# 初始化项目
init: install-tools deps
    @[ -f .env ] || cp .env.example .env

# 同步并整理依赖
sync:
    @echo "Cleaning and syncing dependencies..."
    @{{ GO }} mod tidy
    @echo "Done."

# 下载并整理依赖
deps:
    @{{ GO }} mod download
    @{{ GO }} mod tidy

# 类似 npm ncu：列出直接依赖可升级版本
outdated:
    @sh scripts/mod-gcu.sh list

# 类似 npm ncu -u：更新直接依赖到最新版本
upgrade:
    @sh scripts/mod-gcu.sh up

# 只升级直接依赖的 patch 版本
mod-gcu-up-patch:
    @sh scripts/mod-gcu.sh patch

# 安装开发工具
install-tools:
    @{{ GO }} install github.com/air-verse/air@latest
    @{{ GO }} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @command -v atlas >/dev/null 2>&1 || curl -sSf https://atlasgo.sh | sh -s -- --yes -o /usr/local/bin/atlas
    @{{ GO }} install github.com/swaggo/swag/cmd/swag@latest

# 生成 PostgreSQL 迁移（NAME=xxx）
atlas-diff-postgres name=NAME:
    @{{ ATLAS_MIGRATE }} diff {{ name }} --env postgres

# 应用 PostgreSQL 迁移（依赖 DATABASE_URL）
atlas-apply-postgres:
    @{{ ATLAS_MIGRATE }} apply --env postgres

# 生成 SQLite 迁移（NAME=xxx）
atlas-diff-sqlite name=NAME:
    @{{ ATLAS_MIGRATE }} diff {{ name }} --env sqlite

# 应用 SQLite 迁移
atlas-apply-sqlite:
    @{{ ATLAS_MIGRATE }} apply --env sqlite

# 显示迁移状态（默认 postgres，ENV=sqlite 可切换）
atlas-status env=ENV:
    @{{ ATLAS_MIGRATE }} status --env {{ env }}

# 显示迁移历史（默认 postgres，ENV=sqlite 可切换）
atlas-history env=ENV:
    @{{ ATLAS_MIGRATE }} history --env {{ env }}

# 修复迁移表（默认 postgres，ENV=sqlite 可切换）
atlas-repair env=ENV:
    @{{ ATLAS_MIGRATE }} repair --env {{ env }}

# 重置数据库并重新应用所有迁移（默认 postgres，ENV=sqlite 可切换）
atlas-reset env=ENV:
    @{{ ATLAS_MIGRATE }} reset --env {{ env }}

# 重新生成 PostgreSQL 迁移校验文件（atlas.sum）
atlas-hash-postgres:
    @{{ ATLAS_MIGRATE }} hash --env postgres

# 重新生成 SQLite 迁移校验文件（atlas.sum）
atlas-hash-sqlite:
    @{{ ATLAS_MIGRATE }} hash --env sqlite

# 重新生成迁移校验文件（默认 postgres，ENV=sqlite 可切换）
atlas-hash env=ENV:
    @{{ ATLAS_MIGRATE }} hash --env {{ env }}

# 生成迁移（默认 postgres，NAME=xxx，ENV=sqlite 可切换）
atlas-diff name=NAME env=ENV:
    @{{ ATLAS_MIGRATE }} diff {{ name }} --env {{ env }}

# 应用迁移（默认 postgres，ENV=sqlite 可切换）
atlas-apply env=ENV:
    @{{ ATLAS_MIGRATE }} apply --env {{ env }}

# 检查数据库 schema（默认 postgres，ENV=sqlite 可切换）
atlas-lint env=ENV:
    @{{ ATLAS_SCHEMA }} lint --env {{ env }}

# 检查当前数据库 schema（默认 postgres，ENV=sqlite 可切换）
atlas-inspect env=ENV:
    @{{ ATLAS_SCHEMA }} inspect --env {{ env }}

# 自动从注释生成 Swagger 文档并由官方 Fiber contrib SwaggerUI 展示
docs:
    @swag init --pd --st --parseInternal --packagePrefix lfiber --md docs/md -d cmd/app,internal -g main.go -o docs --ot json
    @python3 internal/support/scripts/reorder_swagger.py docs/swagger.json
    @cp docs/swagger.json docs/openapi.json
    @echo "Documentation generated at docs/openapi.json"

_check-k6:
    @command -v {{ K6 }} >/dev/null 2>&1 || { echo "{{ K6 }} is not installed"; exit 1; }

# 运行根路径 / 的 k6 smoke test（p95 < 1s，错误率 < 1%）
k6-root: _check-k6
    @{{ K6 }} run scripts/k6/scenarios/smoke.js

# 运行根路径 / 的 k6 load test（p95 < 1s，错误率 < 1%）
k6-root-load: _check-k6
    @{{ K6 }} run scripts/k6/scenarios/load.js
