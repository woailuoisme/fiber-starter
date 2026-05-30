#!/bin/bash

# ==============================================================================
# Go Fiber 应用程序编译与打包发布包脚本 (build-release.sh)
# ==============================================================================
#
# 1. 主要功能 (Core Features):
#    - 【环境配置加载】: 自动加载并应用本地的 `.buildconfig` 文件配置项。
#    - 【代码质量门禁】: 打包前强制调用 `just check` (Lint 自动格式化/修复) 及 `just test` (测试套件运行)，
#                       确保只有质量合格、测试通过的代码才会被编译并打包。
#    - 【跨平台编译】: 基于当前 Git 的最新 tag、commit hash 和构建时间，利用 LDFLAGS 注入到主程序中，
#                     交叉编译面向 Linux 平台的 amd64 二进制文件 (压缩大小并移除符号表)。
#    - 【文件结构收集】: 创建隔离的临时目录，收敛二进制文件、运行时配置文件 (configs)、
#                       数据库模型与迁移定义 (database) 及 .env.example 模板文件。
#    - 【归档打包】: 将收集的资源归档并输出为标准 .tar.gz 发布包，便于后续分发与部署。
#
# 2. 命令行用法 (Usage):
#    ./scripts/build-release.sh [staging|production]
#
# 3. 输出产物 (Artifacts):
#    - 构建物路径: deploy/release-<environment>.tar.gz (例如: deploy/release-production.tar.gz)
#    - 包内结构 (Archive Contents):
#        ├── lfiber (主二进制文件)
#        ├── configs/ (所有运行时 yaml 配置目录)
#        ├── database/ (所有数据库迁移脚本与 seed 文件)
#        └── .env.example (环境变量配置模板)
# ==============================================================================

set -euo pipefail
export PATH="/opt/homebrew/bin:/usr/local/bin:$PATH"

if [ -f ".buildconfig" ]; then
    set -a
    # shellcheck source=/dev/null
    . ./.buildconfig
    set +a
fi

# 确保 Go 编译时优先使用本地 go.mod 依赖模式
export GOFLAGS="${GOFLAGS:-} -mod=mod"

APP_MAIN=${APP_MAIN:-./cmd/app}
BUILD_DIR=${BUILD_DIR:-build}
SERVER_BINARY_NAME=${SERVER_BINARY_NAME:-lfiber}
DEPLOY_DIR=${DEPLOY_DIR:-deploy}

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error()   { echo -e "${RED}[ERROR]${NC} $1"; }

show_help() {
    echo "lfiber deployment packager"
    echo ""
    echo "Usage: $0 [environment]"
    echo ""
    echo "Environments:"
    echo "  staging       Build and package for staging"
    echo "  production    Build and package for production"
    echo ""
    echo "Output:"
    echo "  ${DEPLOY_DIR}/release-<env>.tar.gz"
    echo ""
    echo "Examples:"
    echo "  $0 staging"
    echo "  $0 production"
}

# ── 验证环境 ──────────────────────────────────────────────────────────────────

ENVIRONMENT="${1:-}"
case "$ENVIRONMENT" in
    staging|production) ;;
    help|--help|-h) show_help; exit 0 ;;
    "")
        log_error "Environment required."
        show_help
        exit 1
        ;;
    *)
        log_error "Unknown environment: $ENVIRONMENT"
        show_help
        exit 1
        ;;
esac

log_info "Target environment: $ENVIRONMENT"

# ── 代码质量检查 ───────────────────────────────────────────────────────────────

log_info "Running code quality checks..."
just check

log_info "Running tests..."
just test

# ── 编译 ───────────────────────────────────────────────────────────────────────

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-w -s -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

BINARY="$BUILD_DIR/${SERVER_BINARY_NAME}-linux-amd64"
mkdir -p "$BUILD_DIR"

log_info "Building $ENVIRONMENT binary (linux/amd64) — version: $VERSION..."
GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o "$BINARY" "$APP_MAIN"
log_success "Build complete: $BINARY"

# ── 打包 ───────────────────────────────────────────────────────────────────────

RELEASE_DIR="$(mktemp -d)"
trap 'rm -rf "$RELEASE_DIR"' EXIT

cp "$BINARY" "$RELEASE_DIR/$SERVER_BINARY_NAME"
[ -d configs ] && cp -r configs "$RELEASE_DIR/"
[ -d database ] && cp -r database "$RELEASE_DIR/"
[ -f .env.example ] && cp .env.example "$RELEASE_DIR/.env.example"

mkdir -p "$DEPLOY_DIR"
ARCHIVE="$DEPLOY_DIR/release-${ENVIRONMENT}.tar.gz"
tar -czf "$ARCHIVE" -C "$RELEASE_DIR" .

log_success "Release archive: $ARCHIVE"
log_info "Contents:"
tar -tf "$ARCHIVE" | sed 's/^/  /'
