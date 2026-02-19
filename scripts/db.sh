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
        echo "Loading env file: $ENV_FILE"
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
    info "Creating database: $DB_DATABASE"

    # 检查数据库是否已存在
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_DATABASE'" | grep -q 1 && {
        warning "Database $DB_DATABASE already exists"
        return 0
    }

    # 创建数据库
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USERNAME -d postgres -c "CREATE DATABASE \"$DB_DATABASE\";"

    info "Database $DB_DATABASE created successfully!"
}

# 删除数据库
drop_database() {
    warning "About to drop database: $DB_DATABASE"
    read -p "Confirm drop? (yes/no): " confirm

    if [ "$confirm" != "yes" ]; then
        info "Operation cancelled"
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

    info "Database $DB_DATABASE dropped"
}

# 重置数据库
reset_database() {
    info "Resetting database: $DB_DATABASE"
    drop_database
    create_database
    info "Database reset completed"
}

# 运行迁移
migrate() {
    info "Running database migrations..."
	if [ -f "./build/fiber-starter-cli" ]; then
		./build/fiber-starter-cli migrate run
		return
	fi
	go run ./cmd/cli migrate run

    info "Database migrations completed"
}

# 回滚迁移
migrate_rollback() {
    info "Rolling back database migrations..."
	if [ -f "./build/fiber-starter-cli" ]; then
		./build/fiber-starter-cli migrate rollback
		return
	fi
	go run ./cmd/cli migrate rollback

    info "Database migrations rolled back"
}

# 显示帮助信息
show_help() {
    echo "Database management script"
    echo ""
    echo "Usage: ./scripts/db.sh [command]"
    echo ""
    echo "Commands:"
    echo "  create          Create database"
    echo "  drop            Drop database"
    echo "  reset           Reset database (drop and create)"
    echo "  migrate         Run migrations"
    echo "  rollback        Roll back migrations"
    echo "  help            Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  APP_ENV         Environment name (dev, test, production)"
    echo "                  Automatically loads .env.\$APP_ENV"
    echo ""
    echo "Examples:"
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
            error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
