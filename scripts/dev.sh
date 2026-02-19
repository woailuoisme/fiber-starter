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
        log_error "$1 is not installed"
        return 1
    fi
    return 0
}

# 检查项目状态
check_project() {
    log_info "Checking project status..."
    
    # 检查 Go 版本
    if check_command go; then
        log_success "Go version: $(go version)"
    else
        log_error "Go is required. Please install Go first."
        exit 1
    fi
    
    # 检查必要文件
    if [ ! -f "go.mod" ]; then
        log_error "go.mod not found"
        exit 1
    fi
    
	if [ ! -f "cmd/server/main.go" ]; then
		log_error "cmd/server/main.go not found"
        exit 1
    fi
    
    # 检查配置文件
    if [ ! -f ".env" ]; then
        log_warning ".env not found. Creating..."
        cp .env.example .env
        log_success ".env created"
    fi
    
    log_success "Project status check completed"
}

# 安装依赖
install_deps() {
    log_info "Installing project dependencies..."
    go mod download
    go mod tidy
    log_success "Dependencies installed"
}

# 安装开发工具
install_tools() {
    log_info "Installing development tools..."
    
    tools=(
        "github.com/cosmtrek/air"
        "github.com/golangci/golangci-lint/cmd/golangci-lint"
        "github.com/swaggo/swag/cmd/swag"
        "github.com/golang/mock/mockgen"
        "github.com/securecodewarrior/gosec/v2/cmd/gosec"
        "golang.org/x/vuln/cmd/govulncheck"
    )
    
    for tool in "${tools[@]}"; do
        log_info "Installing $tool..."
        go install $tool@latest
    done
    
    log_success "Development tools installed"
}

# 启动开发服务器
start_dev() {
    log_info "Starting development server..."
    
    if check_command air; then
        air
    else
		log_warning "air is not installed. Falling back to go run..."
		go run ./cmd/server
    fi
}

# 运行测试
run_tests() {
    log_info "Running tests..."
    
    # 单元测试
    log_info "Running unit tests..."
    go test -v ./...
    
    # 竞态检测
    log_info "Running race detector..."
    go test -race -v ./...
    
    # 覆盖率
    log_info "Generating coverage report..."
    mkdir -p coverage
    go test -coverprofile=coverage/coverage.out ./...
    go tool cover -html=coverage/coverage.out -o coverage/coverage.html
    
    log_success "Tests completed. Coverage report: coverage/coverage.html"
}

# 代码质量检查
quality_check() {
    log_info "Running code quality checks..."
    
    # 格式化
    log_info "Formatting code..."
    go fmt ./...
    
    # 静态检查
    log_info "Running go vet..."
    go vet ./...
    
    # golangci-lint
    if check_command golangci-lint; then
        log_info "Running golangci-lint..."
        golangci-lint run
    else
        log_warning "golangci-lint is not installed"
    fi
    
    # 安全扫描
    if check_command gosec; then
        log_info "Running security scan..."
        gosec ./...
    else
        log_warning "gosec is not installed"
    fi
    
    log_success "Code quality checks completed"
}

# 构建项目
build_project() {
    log_info "Building project..."
    
    # 创建构建目录
    mkdir -p build
    
    # 构建当前平台
	go build -o build/fiber-starter -v ./cmd/server
    
    # 构建多平台（可选）
    if [ "$1" = "--all" ]; then
        log_info "Building multi-platform binaries..."
        
        # Linux AMD64
		GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o build/fiber-starter-linux-amd64 ./cmd/server
        
        # macOS AMD64
		GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o build/fiber-starter-darwin-amd64 ./cmd/server
        
        # macOS ARM64
		GOOS=darwin GOARCH=arm64 go build -o build/fiber-starter-darwin-arm64 ./cmd/server
        
        # Windows AMD64
		GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o build/fiber-starter-windows-amd64.exe ./cmd/server
    fi
    
    log_success "Build completed"
}

# 数据库操作
db_migrate() {
    log_info "Running database migrations..."
    if [ -f "database/migrations/migrate.go" ]; then
        go run database/migrations/migrate.go up
        log_success "Database migrations completed"
    else
        log_error "Migration file not found"
    fi
}

db_seed() {
    log_info "Running database seeders..."
    if [ -f "database/seeders/seed.go" ]; then
        go run database/seeders/seed.go
        log_success "Database seeding completed"
    else
        log_error "Seeder file not found"
    fi
}

db_reset() {
    log_info "Resetting database..."
    db_migrate_down
    db_migrate
    db_seed
    log_success "Database reset completed"
}

# 生成代码
generate_code() {
    log_info "Generating code..."
    
    # 生成 Go 代码
    go generate ./...
    
    # 生成 Mock 文件
    if check_command mockgen; then
        log_info "Generating mocks..."
        mockgen -source=app/services/*.go -destination=tests/mocks/services_mock.go
    fi
    
    # 生成 API 文档
    if check_command swag; then
        log_info "Generating API docs..."
		swag init -g ./cmd/server/main.go -o docs
    fi
    
    log_success "Code generation completed"
}

# 清理项目
clean_project() {
    log_info "Cleaning project..."
    
    # 清理构建文件
    rm -rf build/
    rm -rf coverage/
    rm -f fiber-starter*
    
    # 清理日志
    rm -f storage/logs/*.log
    
    # 清理临时文件
    find . -name "*.tmp" -delete
    find . -name "*.log" -delete
    
    log_success "Project cleanup completed"
}

# 显示帮助信息
show_help() {
    echo "Fiber Starter development script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Available commands:"
    echo "  check           Check project status"
    echo "  deps            Install dependencies"
    echo "  tools           Install development tools"
    echo "  dev             Start development server"
    echo "  test            Run tests"
    echo "  quality         Run code quality checks"
    echo "  build [options] Build project (--all for multi-platform)"
    echo "  migrate         Run database migrations"
    echo "  seed            Run database seeders"
    echo "  reset           Reset database"
    echo "  generate        Generate code"
    echo "  clean           Clean project"
    echo "  help            Show help"
    echo ""
    echo "Examples:"
    echo "  $0 check"
    echo "  $0 dev"
    echo "  $0 test"
    echo "  $0 build --all"
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
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
