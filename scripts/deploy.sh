#!/bin/bash

# 部署脚本
# 使用方法: ./scripts/deploy.sh [environment]

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

# 默认环境
ENVIRONMENT=${1:-development}

# 检查环境
check_environment() {
    log_info "Checking deployment environment: $ENVIRONMENT"
    
    case $ENVIRONMENT in
        "development"|"staging"|"production")
            log_success "Environment check passed: $ENVIRONMENT"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT"
            log_info "Available environments: development, staging, production"
            exit 1
            ;;
    esac
}

# 准备部署
prepare_deploy() {
    log_info "Preparing deployment..."
    
    # 检查代码状态
    if [ "$ENVIRONMENT" != "development" ]; then
        log_info "Checking Git status..."
        if [ -n "$(git status --porcelain)" ]; then
            log_error "Working directory is not clean. Please commit your changes first."
            exit 1
        fi
    fi
    
    # 运行测试
    log_info "Running tests..."
    ./scripts/dev.sh test
    
    # 代码质量检查
    log_info "Running code quality checks..."
    ./scripts/dev.sh quality
    
    log_success "Deployment preparation completed"
}

# 构建应用
build_app() {
    log_info "Building application..."
    
    # 设置构建信息
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    
    # 构建标志
    LDFLAGS="-w -s -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
    
    case $ENVIRONMENT in
        "development")
			go build -ldflags="$LDFLAGS" -o build/fiber-starter ./cmd/server
            ;;
        "staging")
			GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/fiber-starter-linux-amd64 ./cmd/server
            ;;
        "production")
			GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/fiber-starter-linux-amd64 ./cmd/server
            ;;
    esac
    
    log_success "Build completed"
}

# 部署到不同环境
deploy_to_env() {
    case $ENVIRONMENT in
        "development")
            deploy_development
            ;;
        "staging")
            deploy_staging
            ;;
        "production")
            deploy_production
            ;;
    esac
}

# 开发环境部署
deploy_development() {
    log_info "Deploying to development..."
    
    # 停止现有服务
    if pgrep -f "fiber-starter" > /dev/null; then
        log_info "Stopping existing service..."
        pkill -f "fiber-starter"
        sleep 2
    fi
    
    # 启动新服务
    log_info "Starting development service..."
    nohup ./build/fiber-starter > storage/logs/app.log 2>&1 &
    
    log_success "Development deployment completed"
}

# 预发布环境部署
deploy_staging() {
    log_info "Deploying to staging..."
    
    # 这里可以添加实际的部署逻辑
    # 例如：Docker 部署、Kubernetes 部署等
    
    log_info "Simulating staging deployment..."
    
    # 创建部署包
    mkdir -p deploy/staging
    cp build/fiber-starter-linux-amd64 deploy/staging/fiber-starter
    cp -r config deploy/staging/
    cp -r database deploy/staging/
    cp .env.example deploy/staging/.env
    
    # 创建部署脚本
    cat > deploy/staging/deploy.sh << 'EOF'
#!/bin/bash
# 预发布环境部署脚本

# 停止现有服务
sudo systemctl stop fiber-starter || true

# 备份现有版本
sudo cp /opt/fiber-starter/fiber-starter /opt/fiber-starter/fiber-starter.backup.$(date +%Y%m%d_%H%M%S) || true

# 复制新版本
sudo cp fiber-starter /opt/fiber-starter/
sudo cp -r config /opt/fiber-starter/
sudo cp .env /opt/fiber-starter/

# 设置权限
sudo chmod +x /opt/fiber-starter/fiber-starter

# 启动服务
sudo systemctl start fiber-starter

# 检查服务状态
sudo systemctl status fiber-starter
EOF
    
    chmod +x deploy/staging/deploy.sh
    
    log_success "Staging deployment package is ready"
}

# 生产环境部署
deploy_production() {
    log_info "Deploying to production..."
    
    # 生产环境部署逻辑
    log_warning "Production deployment requires manual confirmation"
    
    # 创建部署包
    mkdir -p deploy/production
    cp build/fiber-starter-linux-amd64 deploy/production/fiber-starter
    cp -r config deploy/production/
    cp -r database deploy/production/
    cp .env.example deploy/production/.env
    
    # 创建部署脚本
    cat > deploy/production/deploy.sh << 'EOF'
#!/bin/bash
# 生产环境部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查权限
if [ "$EUID" -ne 0 ]; then
    log_error "Please run this script with sudo"
    exit 1
fi

# 备份现有版本
if [ -f "/opt/fiber-starter/fiber-starter" ]; then
    log_info "Backing up current version..."
    cp /opt/fiber-starter/fiber-starter /opt/fiber-starter/fiber-starter.backup.$(date +%Y%m%d_%H%M%S)
fi

# 停止服务
log_info "Stopping service..."
systemctl stop fiber-starter || true

# 复制新版本
log_info "Deploying new version..."
cp fiber-starter /opt/fiber-starter/
cp -r config /opt/fiber-starter/
cp .env /opt/fiber-starter/.env.template

# 设置权限
chmod +x /opt/fiber-starter/fiber-starter
chown -R fiber-starter:fiber-starter /opt/fiber-starter/

# 运行数据库迁移
log_info "Running database migrations..."
cd /opt/fiber-starter
sudo -u fiber-starter ./fiber-starter migrate

# 启动服务
log_info "Starting service..."
systemctl start fiber-starter

# 检查服务状态
sleep 5
if systemctl is-active --quiet fiber-starter; then
    log_info "Service started successfully"
else
    log_error "Service failed to start"
    systemctl status fiber-starter
    exit 1
fi

# 健康检查
log_info "Running health check..."
sleep 10
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_info "Health check passed"
else
    log_error "Health check failed"
    exit 1
fi

log_info "Production deployment completed"
EOF
    
    chmod +x deploy/production/deploy.sh
    
    log_success "Production deployment package is ready"
    log_warning "Run deploy/production/deploy.sh manually to perform the actual deployment"
}

# 回滚
rollback() {
    log_info "Rolling back to the previous version..."
    
    case $ENVIRONMENT in
        "development")
            # 开发环境回滚逻辑
            if [ -f "build/fiber-starter.backup" ]; then
                cp build/fiber-starter.backup build/fiber-starter
                log_success "Development rollback completed"
            else
                log_error "Backup file not found"
            fi
            ;;
        "staging"|"production")
            log_info "Please perform rollback manually"
            log_info "Backup files: /opt/fiber-starter/fiber-starter.backup.*"
            ;;
    esac
}

# 显示帮助信息
show_help() {
    echo "Fiber Starter deployment script"
    echo ""
    echo "Usage: $0 [environment] [command]"
    echo ""
    echo "Environments:"
    echo "  development   Development (default)"
    echo "  staging       Staging"
    echo "  production    Production"
    echo ""
    echo "Commands:"
    echo "  deploy        Deploy (default)"
    echo "  rollback      Rollback"
    echo "  help          Show help"
    echo ""
    echo "Examples:"
    echo "  $0"
    echo "  $0 staging deploy"
    echo "  $0 production deploy"
    echo "  $0 staging rollback"
}

# 主函数
main() {
    local command=${2:-deploy}
    
    case $command in
        "deploy")
            check_environment
            prepare_deploy
            build_app
            deploy_to_env
            ;;
        "rollback")
            check_environment
            rollback
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
