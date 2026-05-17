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
        log_success "Task file initialized"
    fi
    
    if [ ! -f "$TASK_LOG" ]; then
        touch "$TASK_LOG"
        log_success "Task log initialized"
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
        log_error "Task title must not be empty"
        return 1
    fi
    
    local id=$(generate_id)
    local created_at=$(get_timestamp)
    local status="pending"
    
    echo "$id|$title|$description|$priority|$status|$created_at|-" >> "$TASK_FILE"
    
    log_success "Task added: $title (ID: $id)"
    log_task "$id" "created" "Task created"
}

# 列出任务
list_tasks() {
    local filter="$1"
    
    init_tasks
    
    if [ ! -s "$TASK_FILE" ]; then
        log_info "No tasks found"
        return 0
    fi
    
    echo ""
    echo -e "${CYAN}=== Task List ===${NC}"
    echo ""
    
    printf "%-10s %-20s %-10s %-10s %-20s\n" "ID" "Title" "Priority" "Status" "Created At"
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
            printf "%-10s   %s\n" "" "Description: $description"
        fi
        
        echo ""
    done < "$TASK_FILE"
}

# 显示任务详情
show_task() {
    local id="$1"
    
    if [ -z "$id" ]; then
        log_error "Please provide a task ID"
        return 1
    fi
    
    local task=$(grep "^$id|" "$TASK_FILE" 2>/dev/null)
    
    if [ -z "$task" ]; then
        log_error "Task not found: $id"
        return 1
    fi
    
    IFS='|' read -r task_id title description priority status created_at completed_at <<< "$task"
    
    echo ""
    echo -e "${CYAN}=== Task Details ===${NC}"
    echo ""
    echo -e "${PURPLE}ID:${NC} $task_id"
    echo -e "${PURPLE}Title:${NC} $title"
    echo -e "${PURPLE}Description:${NC} ${description:-N/A}"
    echo -e "${PURPLE}Priority:${NC} $priority"
    echo -e "${PURPLE}Status:${NC} $status"
    echo -e "${PURPLE}Created At:${NC} $created_at"
    if [ "$completed_at" != "-" ]; then
        echo -e "${PURPLE}Completed At:${NC} $completed_at"
    fi
    echo ""
    
    # 显示任务历史
    echo -e "${CYAN}=== Task History ===${NC}"
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
        log_error "Please provide a task ID and a new status"
        return 1
    fi
    
    # 检查状态是否有效
    case $new_status in
        "pending"|"in_progress"|"completed")
            ;;
        *)
            log_error "Invalid status: $new_status"
            log_info "Valid statuses: pending, in_progress, completed"
            return 1
            ;;
    esac
    
    # 检查任务是否存在
    if ! grep -q "^$id|" "$TASK_FILE"; then
        log_error "Task not found: $id"
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
    
    log_success "Task status updated: $id -> $new_status"
    log_task "$id" "status_changed" "Status changed to $new_status: ${note:-no note}"
}

# 删除任务
delete_task() {
    local id="$1"
    local reason="$2"
    
    if [ -z "$id" ]; then
        log_error "Please provide a task ID"
        return 1
    fi
    
    if ! grep -q "^$id|" "$TASK_FILE"; then
        log_error "Task not found: $id"
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
    
    log_success "Task deleted: $id"
    log_task "$id" "deleted" "Task deleted: ${reason:-no reason}"
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
        log_error "Please provide a search keyword"
        return 1
    fi
    
    echo ""
    echo -e "${CYAN}=== Search Results: $keyword ===${NC}"
    echo ""
    
    local found=false
    while IFS='|' read -r id title description priority status created_at completed_at; do
        if [[ "$title" == *"$keyword"* ]] || [[ "$description" == *"$keyword"* ]]; then
            if [ "$found" = false ]; then
                printf "%-10s %-20s %-10s %-10s %-20s\n" "ID" "Title" "Priority" "Status" "Created At"
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
        log_info "No matching tasks found"
    else
        echo ""
    fi
}

# 显示统计信息
show_stats() {
    init_tasks
    
    if [ ! -s "$TASK_FILE" ]; then
        log_info "No task stats available"
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
    echo -e "${CYAN}=== Task Stats ===${NC}"
    echo ""
    echo -e "${PURPLE}Total:${NC} $total"
    echo ""
    echo -e "${CYAN}By status:${NC}"
    echo -e "  Pending: ${YELLOW}$pending${NC}"
    echo -e "  In progress: ${BLUE}$in_progress${NC}"
    echo -e "  Completed: ${GREEN}$completed${NC}"
    echo ""
    echo -e "${CYAN}By priority:${NC}"
    echo -e "  High: ${RED}$high${NC}"
    echo -e "  Medium: ${YELLOW}$medium${NC}"
    echo -e "  Low: $low"
    echo ""
}

# 清理已完成任务
cleanup_completed() {
    local days="${1:-7}"
    
    log_info "Cleaning up tasks completed more than $days days ago..."
    
    local cutoff_date=$(date -d "$days days ago" '+%Y-%m-%d' 2>/dev/null || date -v-${days}d '+%Y-%m-%d')
    local count=0
    
    temp_file=$(mktemp)
    while IFS='|' read -r id title description priority status created_at completed_at; do
        if [ "$status" = "completed" ] && [ "$completed_at" != "-" ]; then
            local completed_date=${completed_at:0:10}
            if [[ "$completed_date" < "$cutoff_date" ]]; then
                log_task "$id" "auto_deleted" "Auto cleanup: completed more than $days days ago"
                ((count++))
            else
                echo "$id|$title|$description|$priority|$status|$created_at|$completed_at" >> "$temp_file"
            fi
        else
            echo "$id|$title|$description|$priority|$status|$created_at|$completed_at" >> "$temp_file"
        fi
    done < "$TASK_FILE"
    
    mv "$temp_file" "$TASK_FILE"
    
    log_success "Cleaned up $count completed tasks"
}

# 显示帮助信息
show_help() {
    echo "Task manager"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  add <title> [priority] [description]     Add a task"
    echo "  list [filter]                            List tasks (pending/in_progress/completed)"
    echo "  show <id>                                Show task details"
    echo "  update <id> <status> [note]              Update task status"
    echo "  delete <id> [reason]                     Delete a task"
    echo "  search <keyword>                         Search tasks"
    echo "  stats                                    Show stats"
    echo "  cleanup [days]                           Cleanup completed tasks (default: 7)"
    echo "  help                                     Show help"
    echo ""
    echo "Priority: high, medium, low (default: medium)"
    echo "Status: pending, in_progress, completed"
    echo ""
    echo "Examples:"
    echo "  $0 add \"Implement login\" high \"Add JWT authentication\""
    echo "  $0 list pending"
    echo "  $0 update 12345 in_progress \"Start implementation\""
    echo "  $0 show 12345"
    echo "  $0 search \"login\""
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
            log_error "Unknown command: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
