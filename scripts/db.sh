#!/bin/bash

# 数据库管理脚本
# 用法: ./scripts/db.sh [create|drop|reset|migrate]

set -e

# 加载环境变量
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# 如果设置了 APP_ENV，加载对应的环境文件
if [ ! -z "$APP_ENV" ]; then
    ENV_FILE=".env.$APP_ENV"
    if [ -f "$ENV_FILE" ]; then
        echo "加载环境文件: $ENV_FILE"
        export $(cat "$ENV_FILE" | grep -v '^#' | xargs)
    fi
fi

# 数据库连接信息
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USERNAME=${DB_USERNAME:-postgres}
DB_PASSWORD=${DB_PASSWORD:-}
DB_DATABASE=${DB_DATABASE:-fiber_starter}

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印信息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 创建数据库
create_database() {
    info "正在创建数据库: $DB_DATABASE"
    
    # 检查数据库是否已存在
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_DATABASE'" | grep -q 1 && {
        warning "数据库 $DB_DATABASE 已存在"
        return 0
    }
    
    # 创建数据库
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -c "CREATE DATABASE \"$DB_DATABASE\";"
    
    info "数据库 $DB_DATABASE 创建成功！"
}

# 删除数据库
drop_database() {
    warning "即将删除数据库: $DB_DATABASE"
    read -p "确认删除？(yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        info "操作已取消"
        return 0
    fi
    
    # 断开所有连接
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -c "
        SELECT pg_terminate_backend(pg_stat_activity.pid)
        FROM pg_stat_activity
        WHERE pg_stat_activity.datname = '$DB_DATABASE'
        AND pid <> pg_backend_pid();
    " 2>/dev/null || true
    
    # 删除数据库
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -c "DROP DATABASE IF EXISTS \"$DB_DATABASE\";"
    
    info "数据库 $DB_DATABASE 已删除"
}

# 重置数据库
reset_database() {
    info "正在重置数据库: $DB_DATABASE"
    drop_database
    create_database
    info "数据库重置完成"
}

# 运行迁移
migrate() {
    info "正在运行数据库迁移..."
    
    if [ -f "./main" ]; then
        ./main migrate up
    elif [ -f "./fiber-starter" ]; then
        ./fiber-starter migrate up
    else
        go run main.go migrate up
    fi
    
    info "数据库迁移完成"
}

# 回滚迁移
migrate_rollback() {
    info "正在回滚数据库迁移..."
    
    if [ -f "./main" ]; then
        ./main migrate down
    elif [ -f "./fiber-starter" ]; then
        ./fiber-starter migrate down
    else
        go run main.go migrate down
    fi
    
    info "数据库迁移回滚完成"
}

# 显示帮助信息
show_help() {
    echo "数据库管理脚本"
    echo ""
    echo "用法: ./scripts/db.sh [命令]"
    echo ""
    echo "命令:"
    echo "  create          创建数据库"
    echo "  drop            删除数据库"
    echo "  reset           重置数据库（删除并重新创建）"
    echo "  migrate         运行数据库迁移"
    echo "  rollback        回滚数据库迁移"
    echo "  help            显示此帮助信息"
    echo ""
    echo "环境变量:"
    echo "  APP_ENV         环境名称（dev, test, production）"
    echo "                  将自动加载 .env.\$APP_ENV 文件"
    echo ""
    echo "示例:"
    echo "  ./scripts/db.sh create"
    echo "  APP_ENV=dev ./scripts/db.sh reset"
    echo "  APP_ENV=test ./scripts/db.sh migrate"
}

# 主函数
main() {
    case "${1:-help}" in
        create)
            create_database
            ;;
        drop)
            drop_database
            ;;
        reset)
            reset_database
            ;;
        migrate)
            migrate
            ;;
        rollback)
            migrate_rollback
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
