# 定时任务 (Schedule)

类似 Laravel 的定时任务调度系统，基于 `github.com/robfig/cron/v3` 实现。

## 快速开始

### 1. 定义任务

在 `internal/scheduler/tasks.go` 中定义你的任务函数：

```go
package schedule

import (
	"fiber-starter/internal/platform/helpers"
	"go.uber.org/zap"
)

func MyTask() {
	helpers.Info("执行我的任务")
	// 任务逻辑
}
```

### 2. 注册任务

在 `internal/scheduler/kernel.go` 的 `Schedule()` 方法中注册任务：

```go
func (k *Kernel) Schedule() {
	// 每分钟执行
	k.EveryMinute(MyTask)
	
	// 每天凌晨2点执行
	k.DailyAt("02:00", BackupDatabaseTask)
	
	// 每周一早上9点执行
	k.Cron("0 0 9 * * 1", GenerateReportTask)
}
```

### 3. 运行调度器

```bash
# 使用 make 命令（推荐）
make schedule

# 或直接运行
go run ./cmd/cli schedule:run
```

**重要**：
- 调度器独立运行，不会启动 Fiber Web 服务器
- Web 服务器和调度器可以同时运行
- 生产环境建议使用 systemd 或 supervisor 管理调度器进程

## 可用方法

### 基础方法

| 方法 | 说明 | 示例 |
|------|------|------|
| `Cron(spec, cmd)` | 使用 cron 表达式 | `k.Cron("0 0 * * * *", task)` |
| `EverySecond(cmd)` | 每秒执行 | `k.EverySecond(task)` |
| `EveryMinute(cmd)` | 每分钟执行 | `k.EveryMinute(task)` |
| `EveryFiveMinutes(cmd)` | 每5分钟执行 | `k.EveryFiveMinutes(task)` |
| `EveryTenMinutes(cmd)` | 每10分钟执行 | `k.EveryTenMinutes(task)` |
| `EveryFifteenMinutes(cmd)` | 每15分钟执行 | `k.EveryFifteenMinutes(task)` |
| `EveryThirtyMinutes(cmd)` | 每30分钟执行 | `k.EveryThirtyMinutes(task)` |
| `Hourly(cmd)` | 每小时执行 | `k.Hourly(task)` |
| `HourlyAt(minute, cmd)` | 每小时的指定分钟执行 | `k.HourlyAt(30, task)` |
| `Daily(cmd)` | 每天凌晨执行 | `k.Daily(task)` |
| `DailyAt(time, cmd)` | 每天指定时间执行 | `k.DailyAt("13:30", task)` |
| `Weekly(cmd)` | 每周日凌晨执行 | `k.Weekly(task)` |
| `Monthly(cmd)` | 每月1号凌晨执行 | `k.Monthly(task)` |
| `Yearly(cmd)` | 每年1月1号凌晨执行 | `k.Yearly(task)` |

## Cron 表达式格式

支持 6 位 cron 表达式（包含秒）：

```
秒 分 时 日 月 周
*  *  *  *  *  *
```

### 字段说明

| 字段 | 允许值 | 允许的特殊字符 |
|------|--------|----------------|
| 秒 | 0-59 | * / , - |
| 分 | 0-59 | * / , - |
| 时 | 0-23 | * / , - |
| 日 | 1-31 | * / , - ? |
| 月 | 1-12 | * / , - |
| 周 | 0-6 (0=周日) | * / , - ? |

### 特殊字符

- `*` : 匹配所有值
- `/` : 步长值，如 `*/5` 表示每5个单位
- `,` : 列举多个值，如 `1,3,5`
- `-` : 范围，如 `1-5`
- `?` : 不指定值（仅用于日和周）

### 示例

```go
// 每秒执行
k.Cron("* * * * * *", task)

// 每分钟的第30秒执行
k.Cron("30 * * * * *", task)

// 每小时的第30分执行
k.Cron("0 30 * * * *", task)

// 每天凌晨1点执行
k.Cron("0 0 1 * * *", task)

// 每周一到周五的早上9点执行
k.Cron("0 0 9 * * 1-5", task)

// 每月1号和15号的中午12点执行
k.Cron("0 0 12 1,15 * *", task)

// 每5分钟执行
k.Cron("0 */5 * * * *", task)

// 每天早上8点到晚上6点，每小时执行
k.Cron("0 0 8-18 * * *", task)
```

## 实际应用示例

### 1. 数据库备份

```go
func (k *Kernel) Schedule() {
	// 每天凌晨3点备份数据库
	k.DailyAt("03:00", func() {
		helpers.Info("开始备份数据库")
		// 执行备份逻辑
	})
}
```

### 2. 清理过期数据

```go
func (k *Kernel) Schedule() {
	// 每小时清理过期的临时文件
	k.Hourly(func() {
		helpers.Info("清理过期临时文件")
		// 清理逻辑
	})
}
```

### 3. 发送定时报表

```go
func (k *Kernel) Schedule() {
	// 每周一早上9点发送周报
	k.Cron("0 0 9 * * 1", func() {
		helpers.Info("发送周报")
		// 生成并发送报表
	})
}
```

### 4. 同步数据

```go
func (k *Kernel) Schedule() {
	// 每5分钟同步一次数据
	k.EveryFiveMinutes(func() {
		helpers.Info("同步数据")
		// 同步逻辑
	})
}
```

## 生产环境部署

### 使用 systemd (Linux)

创建服务文件 `/etc/systemd/system/app-schedule.service`：

```ini
[Unit]
Description=App Schedule Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/your/app
ExecStart=/path/to/your/app/main schedule:run
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable app-schedule
sudo systemctl start app-schedule
sudo systemctl status app-schedule
```

### 使用 Supervisor

配置文件 `/etc/supervisor/conf.d/app-schedule.conf`：

```ini
[program:app-schedule]
command=/path/to/your/app/main schedule:run
directory=/path/to/your/app
autostart=true
autorestart=true
user=www-data
redirect_stderr=true
stdout_logfile=/var/log/app-schedule.log
```

### 使用 Docker

```dockerfile
# 在 Dockerfile 中添加
CMD ["./main", "schedule:run"]
```

或使用 docker-compose：

```yaml
services:
  schedule:
    build: .
    command: ./main schedule:run
    restart: always
```

## 注意事项

1. **时区问题**：确保服务器时区设置正确
2. **并发执行**：如果任务执行时间超过调度间隔，可能会并发执行
3. **错误处理**：在任务中添加适当的错误处理和日志记录
4. **资源占用**：注意任务的资源消耗，避免影响主应用
5. **单实例运行**：建议只运行一个调度器实例，避免重复执行

## 与 Laravel Schedule 的对比

| Laravel | Go Schedule | 说明 |
|---------|-------------|------|
| `->everyMinute()` | `EveryMinute()` | 每分钟 |
| `->everyFiveMinutes()` | `EveryFiveMinutes()` | 每5分钟 |
| `->hourly()` | `Hourly()` | 每小时 |
| `->daily()` | `Daily()` | 每天 |
| `->dailyAt('13:00')` | `DailyAt("13:00")` | 每天指定时间 |
| `->weekly()` | `Weekly()` | 每周 |
| `->monthly()` | `Monthly()` | 每月 |
| `->cron('* * * * *')` | `Cron("0 * * * * *")` | Cron 表达式（注意多了秒） |

主要区别：
- Go 版本支持秒级定时（6位表达式）
- Laravel 是5位表达式（不含秒）
- Go 版本需要手动运行调度器进程
- Laravel 通过系统 cron 调用 `schedule:run`
