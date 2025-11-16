#!/bin/bash

# 任务管理工具
# 使用方法: ./scripts/task.sh [command] [options]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
TASK_FILE=".tasks"
TASK_LOG=".task_log"

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

# 初始化任务文件
init_tasks() {
    if [ ! -f "$TASK_FILE" ]; then
        touch "$TASK_FILE"
        log_success "任务文件已初始化"
    fi
    
    if [ ! -f "$TASK_LOG" ]; then
        touch "$TASK_LOG"
        log_success "任务日志已初始化"
    fi
}

# 生成任务ID
generate_id() {
    echo $(date +%s%N | tail -c 10)
}

# 获取当前时间戳
get_timestamp() {
    date '+%Y-%m-%d %H:%M:%S'
}

# 添加任务
add_task() {
    local title="$1"
    local priority="${2:-medium}"
    local description="$3"
    
    if [ -z "$title" ]; then
        log_error "任务标题不能为空"
        return 1
    fi
    
    local id=$(generate_id)
    local created_at=$(get_timestamp)
    local status="pending"
    
    echo "$id|$title|$description|$priority|$status|$created_at|-" >> "$TASK_FILE"
    
    log_success "任务已添加: $title (ID: $id)"
    log_task "$id" "created" "任务创建"
}

# 列出任务
list_tasks() {
    local filter="$1"
    
    init_tasks
    
    if [ ! -s "$TASK_FILE" ]; then
        log_info "暂无任务"
        return 0
    fi
    
    echo ""
    echo -e "${CYAN}=== 任务列表 ===${NC}"
    echo ""
    
    printf "%-10s %-20s %-10s %-10s %-20s\n" "ID" "标题" "优先级" "状态" "创建时间"
    printf "%-10s %-20s %-10s %-10s %-20s\n" "----" "----" "----" "----" "----"
    
    while IFS='|' read -r id title description priority status created_at completed_at; do
        if [ -n "$filter" ] && [ "$status" != "$filter" ]; then
            continue
        fi
        
        # 截断长标题
        if [ ${#title} -gt 18 ]; then
            title="${title:0:15}..."
        fi
        
        # 优先级颜色
        case $priority in
            "high")
                priority_color="${RED}$priority${NC}"
                ;;
            "medium")
                priority_color="${YELLOW}$priority${NC}"
                ;;
            "low")
                priority_color="$priority"
                ;;
        esac
        
        # 状态颜色
        case $status in
            "pending")
                status_color="${YELLOW}$status${NC}"
                ;;
            "in_progress")
                status_color="${BLUE}$status${NC}"
                ;;
            "completed")
                status_color="${GREEN}$status${NC}"
                ;;
        esac
        
        printf "%-10s %-20s %-10s %-10s %-20s\n" "$id" "$title" "$priority_color" "$status_color" "${created_at:0:16}"
        
        # 显示描述
        if [ -n "$description" ]; then
            printf "%-10s   %s\n" "" "描述: $description"
        fi
        
        echo ""
    done < "$TASK_FILE"
}

# 显示任务详情
show_task() {
    local id="$1"
    
    if [ -z "$id" ]; then
        log_error "请提供任务ID"
        return 1
    fi
    
    local task=$(grep "^$id|" "$TASK_FILE" 2>/dev/null)
    
    if [ -z "$task" ]; then
        log_error "任务不存在: $id"
        return 1
    fi
    
    IFS='|' read -r task_id title description priority status created_at completed_at <<< "$task"
    
    echo ""
    echo -e "${CYAN}=== 任务详情 ===${NC}"
    echo ""
    echo -e "${PURPLE}ID:${NC} $task_id"
    echo -e "${PURPLE}标题:${NC} $title"
    echo -e "${PURPLE}描述:${NC} ${description:-无}"
    echo -e "${PURPLE}优先级:${NC} $priority"
    echo -e "${PURPLE}状态:${NC} $status"
    echo -e "${PURPLE}创建时间:${NC} $created_at"
    if [ "$completed_at" != "-" ]; then
        echo -e "${PURPLE}完成时间:${NC} $completed_at"
    fi
    echo ""
    
    # 显示任务历史
    echo -e "${CYAN}=== 任务历史 ===${NC}"
    grep "^$id|" "$TASK_LOG" 2>/dev/null | while IFS='|' read -r log_id log_action log_note log_time; do
        echo -e "${BLUE}$log_time${NC} - ${GREEN}$log_action${NC}: $log_note"
    done
    echo ""
}

# 更新任务状态
update_task() {
    local id="$1"
    local new_status="$2"
    local note="$3"
    
    if [ -z "$id" ] || [ -z "$new_status" ]; then
        log_error "请提供任务ID和新状态"
        return 1
    fi
    
    # 检查状态是否有效
    case $new_status in
        "pending"|"in_progress"|"completed")
            ;;
        *)
            log_error "无效的状态: $new_status"
            log_info "可用状态: pending, in_progress, completed"
            return 1
            ;;
    esac
    
    # 检查任务是否存在
    if ! grep -q "^$id|" "$TASK_FILE"; then
        log_error "任务不存在: $id"
        return 1
    fi
    
    # 更新任务状态
    local completed_at="-"
    if [ "$new_status" = "completed" ]; then
        completed_at=$(get_timestamp)
    fi
    
    # 使用临时文件更新
    temp_file=$(mktemp)
    while IFS='|' read -r task_id title description priority status created_at old_completed_at; do
        if [ "$task_id" = "$id" ]; then
            echo "$task_id|$title|$description|$priority|$new_status|$created_at|$completed_at" >> "$temp_file"
        else
            echo "$task_id|$title|$description|$priority|$status|$created_at|$old_completed_at" >> "$temp_file"
        fi
    done < "$TASK_FILE"
    
    mv "$temp_file" "$TASK_FILE"
    
    log_success "任务状态已更新: $id -> $new_status"
    log_task "$id" "status_changed" "状态变更为 $new_status: ${note:-无备注}"
}

# 删除任务
delete_task() {
    local id="$1"
    local reason="$2"
    
    if [ -z "$id" ]; then
        log_error "请提供任务ID"
        return 1
    fi
    
    if ! grep -q "^$id|" "$TASK_FILE"; then
        log_error "任务不存在: $id"
        return 1
    fi
    
    # 使用临时文件删除任务
    temp_file=$(mktemp)
    while IFS='|' read -r task_id; do
        if [ "$task_id" != "$id" ]; then
            grep "^$task_id|" "$TASK_FILE" >> "$temp_file"
        fi
    done < <(cut -d'|' -f1 "$TASK_FILE")
    
    mv "$temp_file" "$TASK_FILE"
    
    log_success "任务已删除: $id"
    log_task "$id" "deleted" "任务删除: ${reason:-无原因}"
}

# 记录任务日志
log_task() {
    local id="$1"
    local action="$2"
    local note="$3"
    local timestamp=$(get_timestamp)
    
    echo "$id|$action|$note|$timestamp" >> "$TASK_LOG"
}

# 搜索任务
search_tasks() {
    local keyword="$1"
    
    if [ -z "$keyword" ]; then
        log_error "请提供搜索关键词"
        return 1
    fi
    
    echo ""
    echo -e "${CYAN}=== 搜索结果: $keyword ===${NC}"
    echo ""
    
    local found=false
    while IFS='|' read -r id title description priority status created_at completed_at; do
        if [[ "$title" == *"$keyword"* ]] || [[ "$description" == *"$keyword"* ]]; then
            if [ "$found" = false ]; then
                printf "%-10s %-20s %-10s %-10s %-20s\n" "ID" "标题" "优先级" "状态" "创建时间"
                printf "%-10s %-20s %-10s %-10s %-20s\n" "----" "----" "----" "----" "----"
                found=true
            fi
            
            # 截断长标题
            if [ ${#title} -gt 18 ]; then
                title="${title:0:15}..."
            fi
            
            printf "%-10s %-20s %-10s %-10s %-20s\n" "$id" "$title" "$priority" "$status" "${created_at:0:16}"
        fi
    done < "$TASK_FILE"
    
    if [ "$found" = false ]; then
        log_info "未找到匹配的任务"
    else
        echo ""
    fi
}

# 显示统计信息
show_stats() {
    init_tasks
    
    if [ ! -s "$TASK_FILE" ]; then
        log_info "暂无任务统计"
        return 0
    fi
    
    local total=$(wc -l < "$TASK_FILE")
    local pending=$(grep -c "|pending|" "$TASK_FILE" 2>/dev/null || echo 0)
    local in_progress=$(grep -c "|in_progress|" "$TASK_FILE" 2>/dev/null || echo 0)
    local completed=$(grep -c "|completed|" "$TASK_FILE" 2>/dev/null || echo 0)
    local high=$(grep -c "|high|" "$TASK_FILE" 2>/dev/null || echo 0)
    local medium=$(grep -c "|medium|" "$TASK_FILE" 2>/dev/null || echo 0)
    local low=$(grep -c "|low|" "$TASK_FILE" 2>/dev/null || echo 0)
    
    echo ""
    echo -e "${CYAN}=== 任务统计 ===${NC}"
    echo ""
    echo -e "${PURPLE}总任务数:${NC} $total"
    echo ""
    echo -e "${CYAN}按状态统计:${NC}"
    echo -e "  待处理: ${YELLOW}$pending${NC}"
    echo -e "  进行中: ${BLUE}$in_progress${NC}"
    echo -e "  已完成: ${GREEN}$completed${NC}"
    echo ""
    echo -e "${CYAN}按优先级统计:${NC}"
    echo -e "  高优先级: ${RED}$high${NC}"
    echo -e "  中优先级: ${YELLOW}$medium${NC}"
    echo -e "  低优先级: $low"
    echo ""
}

# 清理已完成任务
cleanup_completed() {
    local days="${1:-7}"
    
    log_info "清理 $days 天前已完成的任务..."
    
    local cutoff_date=$(date -d "$days days ago" '+%Y-%m-%d' 2>/dev/null || date -v-${days}d '+%Y-%m-%d')
    local count=0
    
    temp_file=$(mktemp)
    while IFS='|' read -r id title description priority status created_at completed_at; do
        if [ "$status" = "completed" ] && [ "$completed_at" != "-" ]; then
            local completed_date=${completed_at:0:10}
            if [[ "$completed_date" < "$cutoff_date" ]]; then
                log_task "$id" "auto_deleted" "自动清理: $days 天前完成"
                ((count++))
            else
                echo "$id|$title|$description|$priority|$status|$created_at|$completed_at" >> "$temp_file"
            fi
        else
            echo "$id|$title|$description|$priority|$status|$created_at|$completed_at" >> "$temp_file"
        fi
    done < "$TASK_FILE"
    
    mv "$temp_file" "$TASK_FILE"
    
    log_success "已清理 $count 个已完成的任务"
}

# 显示帮助信息
show_help() {
    echo "任务管理工具"
    echo ""
    echo "使用方法: $0 [command] [options]"
    echo ""
    echo "命令:"
    echo "  add <title> [priority] [description]     添加任务"
    echo "  list [filter]                            列出任务 (filter: pending/in_progress/completed)"
    echo "  show <id>                                 显示任务详情"
    echo "  update <id> <status> [note]              更新任务状态"
    echo "  delete <id> [reason]                      删除任务"
    echo "  search <keyword>                          搜索任务"
    echo "  stats                                     显示统计信息"
    echo "  cleanup [days]                            清理已完成任务 (默认7天)"
    echo "  help                                      显示帮助信息"
    echo ""
    echo "优先级: high, medium, low (默认: medium)"
    echo "状态: pending, in_progress, completed"
    echo ""
    echo "示例:"
    echo "  $0 add \"实现用户登录\" high \"添加JWT认证功能\""
    echo "  $0 list pending"
    echo "  $0 update 12345 in_progress \"开始开发\""
    echo "  $0 show 12345"
    echo "  $0 search \"登录\""
    echo "  $0 stats"
}

# 主函数
main() {
    init_tasks
    
    case "$1" in
        "add")
            add_task "$2" "$3" "$4"
            ;;
        "list")
            list_tasks "$2"
            ;;
        "show")
            show_task "$2"
            ;;
        "update")
            update_task "$2" "$3" "$4"
            ;;
        "delete")
            delete_task "$2" "$3"
            ;;
        "search")
            search_tasks "$2"
            ;;
        "stats")
            show_stats
            ;;
        "cleanup")
            cleanup_completed "$2"
            ;;
        "help"|"--help"|"-h"|"")
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