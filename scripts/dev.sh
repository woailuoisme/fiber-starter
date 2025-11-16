#!/bin/bash

# Fiber Starter 开发脚本
# 使用方法: ./scripts/dev.sh [command]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装"
        return 1
    fi
    return 0
}

# 检查项目状态
check_project() {
    log_info "检查项目状态..."
    
    # 检查 Go 版本
    if check_command go; then
        log_success "Go 版本: $(go version)"
    else
        log_error "请先安装 Go"
        exit 1
    fi
    
    # 检查必要文件
    if [ ! -f "go.mod" ]; then
        log_error "go.mod 文件不存在"
        exit 1
    fi
    
    if [ ! -f "main.go" ]; then
        log_error "main.go 文件不存在"
        exit 1
    fi
    
    # 检查配置文件
    if [ ! -f ".env" ]; then
        log_warning ".env 文件不存在，正在创建..."
        cp .env.example .env
        log_success "已创建 .env 文件"
    fi
    
    log_success "项目状态检查完成"
}

# 安装依赖
install_deps() {
    log_info "安装项目依赖..."
    go mod download
    go mod tidy
    log_success "依赖安装完成"
}

# 安装开发工具
install_tools() {
    log_info "安装开发工具..."
    
    tools=(
        "github.com/cosmtrek/air"
        "github.com/golangci/golangci-lint/cmd/golangci-lint"
        "github.com/swaggo/swag/cmd/swag"
        "github.com/golang/mock/mockgen"
        "github.com/securecodewarrior/gosec/v2/cmd/gosec"
        "golang.org/x/vuln/cmd/govulncheck"
    )
    
    for tool in "${tools[@]}"; do
        log_info "安装 $tool..."
        go install $tool@latest
    done
    
    log_success "开发工具安装完成"
}

# 启动开发服务器
start_dev() {
    log_info "启动开发服务器..."
    
    if check_command air; then
        air
    else
        log_warning "air 未安装，使用 go run 启动..."
        go run main.go
    fi
}

# 运行测试
run_tests() {
    log_info "运行测试..."
    
    # 单元测试
    log_info "运行单元测试..."
    go test -v ./...
    
    # 竞态检测
    log_info "运行竞态检测..."
    go test -race -v ./...
    
    # 覆盖率
    log_info "生成覆盖率报告..."
    mkdir -p coverage
    go test -coverprofile=coverage/coverage.out ./...
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    
    log_success "测试完成，覆盖率报告: coverage/coverage.html"
}

# 代码质量检查
quality_check() {
    log_info "运行代码质量检查..."
    
    # 格式化
    log_info "格式化代码..."
    go fmt ./...
    
    # 静态检查
    log_info "运行 go vet..."
    go vet ./...
    
    # golangci-lint
    if check_command golangci-lint; then
        log_info "运行 golangci-lint..."
        golangci-lint run
    else
        log_warning "golangci-lint 未安装"
    fi
    
    # 安全扫描
    if check_command gosec; then
        log_info "运行安全扫描..."
        gosec ./...
    else
        log_warning "gosec 未安装"
    fi
    
    log_success "代码质量检查完成"
}

# 构建项目
build_project() {
    log_info "构建项目..."
    
    # 创建构建目录
    mkdir -p build
    
    # 构建当前平台
    go build -o build/fiber-starter -v main.go
    
    # 构建多平台（可选）
    if [ "$1" = "--all" ]; then
        log_info "构建多平台版本..."
        
        # Linux AMD64
        GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o build/fiber-starter-linux-amd64 main.go
        
        # macOS AMD64
        GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o build/fiber-starter-darwin-amd64 main.go
        
        # macOS ARM64
        GOOS=darwin GOARCH=arm64 go build -o build/fiber-starter-darwin-arm64 main.go
        
        # Windows AMD64
        GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o build/fiber-starter-windows-amd64.exe main.go
    fi
    
    log_success "构建完成"
}

# 数据库操作
db_migrate() {
    log_info "运行数据库迁移..."
    if [ -f "database/migrations/migrate.go" ]; then
        go run database/migrations/migrate.go up
        log_success "数据库迁移完成"
    else
        log_error "迁移文件不存在"
    fi
}

db_seed() {
    log_info "运行数据库种子..."
    if [ -f "database/seeders/seed.go" ]; then
        go run database/seeders/seed.go
        log_success "数据库种子完成"
    else
        log_error "种子文件不存在"
    fi
}

db_reset() {
    log_info "重置数据库..."
    db_migrate_down
    db_migrate
    db_seed
    log_success "数据库重置完成"
}

# 生成代码
generate_code() {
    log_info "生成代码..."
    
    # 生成 Go 代码
    go generate ./...
    
    # 生成 Mock 文件
    if check_command mockgen; then
        log_info "生成 Mock 文件..."
        mockgen -source=app/services/*.go -destination=tests/mocks/services_mock.go
    fi
    
    # 生成 API 文档
    if check_command swag; then
        log_info "生成 API 文档..."
        swag init -g main.go -o docs
    fi
    
    log_success "代码生成完成"
}

# 清理项目
clean_project() {
    log_info "清理项目..."
    
    # 清理构建文件
    rm -rf build/
    rm -rf coverage/
    rm -f fiber-starter*
    
    # 清理日志
    rm -f storage/logs/*.log
    
    # 清理临时文件
    find . -name "*.tmp" -delete
    find . -name "*.log" -delete
    
    log_success "项目清理完成"
}

# 显示帮助信息
show_help() {
    echo "Fiber Starter 开发脚本"
    echo ""
    echo "使用方法: $0 [command]"
    echo ""
    echo "可用命令:"
    echo "  check           检查项目状态"
    echo "  deps            安装项目依赖"
    echo "  tools           安装开发工具"
    echo "  dev             启动开发服务器"
    echo "  test            运行测试"
    echo "  quality         代码质量检查"
    echo "  build [options] 构建项目 (使用 --all 构建多平台版本)"
    echo "  migrate         运行数据库迁移"
    echo "  seed            运行数据库种子"
    echo "  reset           重置数据库"
    echo "  generate        生成代码"
    echo "  clean           清理项目"
    echo "  help            显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 check        # 检查项目状态"
    echo "  $0 dev          # 启动开发服务器"
    echo "  $0 test         # 运行测试"
    echo "  $0 build --all  # 构建多平台版本"
}

# 主函数
main() {
    case "$1" in
        "check")
            check_project
            ;;
        "deps")
            install_deps
            ;;
        "tools")
            install_tools
            ;;
        "dev")
            check_project
            start_dev
            ;;
        "test")
            run_tests
            ;;
        "quality")
            quality_check
            ;;
        "build")
            build_project "$2"
            ;;
        "migrate")
            db_migrate
            ;;
        "seed")
            db_seed
            ;;
        "reset")
            db_reset
            ;;
        "generate")
            generate_code
            ;;
        "clean")
            clean_project
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"