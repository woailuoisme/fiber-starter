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
    log_info "检查部署环境: $ENVIRONMENT"
    
    case $ENVIRONMENT in
        "development"|"staging"|"production")
            log_success "环境检查通过: $ENVIRONMENT"
            ;;
        *)
            log_error "无效的环境: $ENVIRONMENT"
            log_info "可用环境: development, staging, production"
            exit 1
            ;;
    esac
}

# 准备部署
prepare_deploy() {
    log_info "准备部署..."
    
    # 检查代码状态
    if [ "$ENVIRONMENT" != "development" ]; then
        log_info "检查 Git 状态..."
        if [ -n "$(git status --porcelain)" ]; then
            log_error "工作目录不干净，请先提交更改"
            exit 1
        fi
    fi
    
    # 运行测试
    log_info "运行测试..."
    ./scripts/dev.sh test
    
    # 代码质量检查
    log_info "运行代码质量检查..."
    ./scripts/dev.sh quality
    
    log_success "部署准备完成"
}

# 构建应用
build_app() {
    log_info "构建应用..."
    
    # 设置构建信息
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    
    # 构建标志
    LDFLAGS="-w -s -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"
    
    case $ENVIRONMENT in
        "development")
            go build -ldflags="$LDFLAGS" -o build/fiber-starter main.go
            ;;
        "staging")
            GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/fiber-starter-linux-amd64 main.go
            ;;
        "production")
            GOOS=linux GOARCH=amd64 go build -ldflags="$LDFLAGS" -o build/fiber-starter-linux-amd64 main.go
            ;;
    esac
    
    log_success "应用构建完成"
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
    log_info "部署到开发环境..."
    
    # 停止现有服务
    if pgrep -f "fiber-starter" > /dev/null; then
        log_info "停止现有服务..."
        pkill -f "fiber-starter"
        sleep 2
    fi
    
    # 启动新服务
    log_info "启动开发服务..."
    nohup ./build/fiber-starter > storage/logs/app.log 2>&1 &
    
    log_success "开发环境部署完成"
}

# 预发布环境部署
deploy_staging() {
    log_info "部署到预发布环境..."
    
    # 这里可以添加实际的部署逻辑
    # 例如：Docker 部署、Kubernetes 部署等
    
    log_info "模拟预发布部署..."
    
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
    
    log_success "预发布环境部署包准备完成"
}

# 生产环境部署
deploy_production() {
    log_info "部署到生产环境..."
    
    # 生产环境部署逻辑
    log_warning "生产环境部署需要人工确认"
    
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
    log_error "请使用 sudo 运行此脚本"
    exit 1
fi

# 备份现有版本
if [ -f "/opt/fiber-starter/fiber-starter" ]; then
    log_info "备份现有版本..."
    cp /opt/fiber-starter/fiber-starter /opt/fiber-starter/fiber-starter.backup.$(date +%Y%m%d_%H%M%S)
fi

# 停止服务
log_info "停止服务..."
systemctl stop fiber-starter || true

# 复制新版本
log_info "部署新版本..."
cp fiber-starter /opt/fiber-starter/
cp -r config /opt/fiber-starter/
cp .env /opt/fiber-starter/.env.template

# 设置权限
chmod +x /opt/fiber-starter/fiber-starter
chown -R fiber-starter:fiber-starter /opt/fiber-starter/

# 运行数据库迁移
log_info "运行数据库迁移..."
cd /opt/fiber-starter
sudo -u fiber-starter ./fiber-starter migrate

# 启动服务
log_info "启动服务..."
systemctl start fiber-starter

# 检查服务状态
sleep 5
if systemctl is-active --quiet fiber-starter; then
    log_info "服务启动成功"
else
    log_error "服务启动失败"
    systemctl status fiber-starter
    exit 1
fi

# 健康检查
log_info "进行健康检查..."
sleep 10
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_info "健康检查通过"
else
    log_error "健康检查失败"
    exit 1
fi

log_info "生产环境部署完成"
EOF
    
    chmod +x deploy/production/deploy.sh
    
    log_success "生产环境部署包准备完成"
    log_warning "请手动执行 deploy/production/deploy.sh 进行实际部署"
}

# 回滚
rollback() {
    log_info "回滚到上一个版本..."
    
    case $ENVIRONMENT in
        "development")
            # 开发环境回滚逻辑
            if [ -f "build/fiber-starter.backup" ]; then
                cp build/fiber-starter.backup build/fiber-starter
                log_success "开发环境回滚完成"
            else
                log_error "没有找到备份文件"
            fi
            ;;
        "staging"|"production")
            log_info "请手动执行回滚操作"
            log_info "备份文件位置: /opt/fiber-starter/fiber-starter.backup.*"
            ;;
    esac
}

# 显示帮助信息
show_help() {
    echo "Fiber Starter 部署脚本"
    echo ""
    echo "使用方法: $0 [environment] [command]"
    echo ""
    echo "环境:"
    echo "  development   开发环境 (默认)"
    echo "  staging       预发布环境"
    echo "  production    生产环境"
    echo ""
    echo "命令:"
    echo "  deploy        部署 (默认)"
    echo "  rollback      回滚"
    echo "  help          显示帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                    # 部署到开发环境"
    echo "  $0 staging deploy     # 部署到预发布环境"
    echo "  $0 production deploy  # 部署到生产环境"
    echo "  $0 staging rollback   # 预发布环境回滚"
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
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"