# Schedule 定时任务

类似 Laravel Schedule 的定时任务系统。

## 快速开始

### 1. 创建任务

在 `tasks.go` 中定义你的任务：

```go
package schedule

import (
	"fiber-starter/app/helpers"
	"go.uber.org/zap"
)

func MyCustomTask() {
	helpers.Info("执行自定义任务")
	
	// 你的业务逻辑
	// 例如：清理过期数据、发送通知、生成报表等
}
```

### 2. 注册任务

在 `kernel.go` 的 `Schedule()` 方法中注册：

```go
func (k *Kernel) Schedule() {
	// 每分钟执行
	k.EveryMinute(MyCustomTask)
	
	// 每天凌晨3点执行
	k.DailyAt("03:00", BackupDatabaseTask)
	
	// 使用 cron 表达式：每周一到周五早上9点
	k.Cron("0 0 9 * * 1-5", GenerateReportTask)
}
```

### 3. 运行调度器

```bash
# 方式1：使用 make 命令（推荐）
make schedule

# 方式2：直接运行
go run cli.go schedule:run

# 方式3：编译后运行
go build -o fiber-starter cli.go
./fiber-starter schedule:run
```

**注意**：调度器是独立运行的，不会启动 Fiber Web 服务器。

## 常用方法

```go
// 时间间隔
k.EverySecond(task)          // 每秒
k.EveryMinute(task)          // 每分钟
k.EveryFiveMinutes(task)     // 每5分钟
k.EveryTenMinutes(task)      // 每10分钟
k.EveryFifteenMinutes(task)  // 每15分钟
k.EveryThirtyMinutes(task)   // 每30分钟
k.Hourly(task)               // 每小时
k.HourlyAt(30, task)         // 每小时的30分

// 每天
k.Daily(task)                // 每天凌晨
k.DailyAt("13:30", task)     // 每天13:30

// 每周/月/年
k.Weekly(task)               // 每周日凌晨
k.Monthly(task)              // 每月1号凌晨
k.Yearly(task)               // 每年1月1号凌晨

// 自定义 cron 表达式
k.Cron("0 */5 * * * *", task) // 每5分钟
```

## Cron 表达式

格式：`秒 分 时 日 月 周`

```go
// 每天早上8点到晚上6点，每小时执行
k.Cron("0 0 8-18 * * *", task)

// 每月1号和15号中午12点
k.Cron("0 0 12 1,15 * *", task)

// 每周一到周五早上9点
k.Cron("0 0 9 * * 1-5", task)
```

## 实际应用示例

### 数据库备份

```go
func BackupDatabaseTask() {
	helpers.Info("开始备份数据库")
	
	// 执行备份命令
	// cmd := exec.Command("mysqldump", "-u", "user", "-p", "database")
	// ...
	
	helpers.Info("数据库备份完成")
}

// 注册：每天凌晨3点执行
k.DailyAt("03:00", BackupDatabaseTask)
```

### 清理过期数据

```go
func CleanupTask() {
	helpers.Info("清理过期数据")
	
	// 清理7天前的临时文件
	// 清理过期的会话
	// 清理过期的缓存
	
	helpers.Info("清理完成")
}

// 注册：每小时执行
k.Hourly(CleanupTask)
```

### 发送定时报表

```go
func GenerateReportTask() {
	helpers.Info("生成日报")
	
	// 统计昨天的数据
	// 生成报表
	// 发送邮件
	
	helpers.Info("报表已发送")
}

// 注册：每天早上9点执行
k.DailyAt("09:00", GenerateReportTask)
```

### 同步第三方数据

```go
func SyncDataTask() {
	helpers.Info("同步第三方数据")
	
	// 调用第三方 API
	// 更新本地数据库
	
	helpers.Info("数据同步完成")
}

// 注册：每5分钟执行
k.EveryFiveMinutes(SyncDataTask)
```

## 生产环境部署

### 使用 systemd (推荐)

创建 `/etc/systemd/system/app-schedule.service`：

```ini
[Unit]
Description=App Schedule Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/app
ExecStart=/var/www/app/fiber-starter schedule:run
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动：

```bash
sudo systemctl enable app-schedule
sudo systemctl start app-schedule
sudo systemctl status app-schedule
```

### 使用 Docker

```yaml
# docker-compose.yml
services:
  app:
    build: .
    command: ./fiber-starter
    
  schedule:
    build: .
    command: ./fiber-starter schedule:run
    restart: always
```

## 注意事项

1. **时区**：确保服务器时区正确
2. **日志**：任务中添加适当的日志记录
3. **错误处理**：捕获并记录错误，避免任务崩溃
4. **性能**：注意任务执行时间，避免阻塞
5. **单实例**：只运行一个调度器实例

## 调试

```bash
# 查看日志
tail -f storage/logs/app.log

# 测试单个任务
# 在 main.go 中临时调用任务函数
```

## 与 Laravel 对比

| Laravel | Go | 说明 |
|---------|-----|------|
| `$schedule->everyMinute()` | `k.EveryMinute()` | 每分钟 |
| `$schedule->hourly()` | `k.Hourly()` | 每小时 |
| `$schedule->daily()` | `k.Daily()` | 每天 |
| `$schedule->dailyAt('13:00')` | `k.DailyAt("13:00")` | 指定时间 |
| `$schedule->cron('* * * * *')` | `k.Cron("0 * * * * *")` | Cron 表达式 |

主要区别：
- Go 支持秒级定时（6位表达式）
- Laravel 是5位表达式
- Go 需要独立运行调度器进程
