# 定时任务快速开始

## 1. 快速测试

取消注释 `kernel.go` 中的测试任务：

```go
// 🧪 快速测试：每10秒执行一次（取消注释以测试）
k.Cron("*/10 * * * * *", TestScheduleTask)
```

运行调度器：

```bash
make schedule
```

你会看到每10秒输出一次日志：

```
🎯 定时任务测试 time=2024-12-17 20:50:00
🎯 定时任务测试 time=2024-12-17 20:50:10
🎯 定时任务测试 time=2024-12-17 20:50:20
```

按 `Ctrl+C` 停止调度器。

## 2. 创建你的第一个任务

### 步骤 1：在 `tasks.go` 中定义任务

```go
func MyFirstTask() {
    helpers.Info("我的第一个定时任务")
    // 你的业务逻辑
}
```

### 步骤 2：在 `kernel.go` 中注册任务

```go
func (k *Kernel) Schedule() {
    // 每分钟执行
    k.EveryMinute(MyFirstTask)
}
```

### 步骤 3：运行调度器

```bash
make schedule
```

## 3. 常用场景

### 每天凌晨备份数据库

```go
func BackupDatabase() {
    helpers.Info("开始备份数据库")
    // 备份逻辑
}

// 注册：每天凌晨3点
k.DailyAt("03:00", BackupDatabase)
```

### 每小时清理临时文件

```go
func CleanupTemp() {
    helpers.Info("清理临时文件")
    // 清理逻辑
}

// 注册：每小时
k.Hourly(CleanupTemp)
```

### 工作日早上发送报表

```go
func SendDailyReport() {
    helpers.Info("发送日报")
    // 发送逻辑
}

// 注册：周一到周五早上9点
k.Cron("0 0 9 * * 1-5", SendDailyReport)
```

## 4. 生产环境部署

### 使用 systemd

```bash
# 创建服务文件
sudo nano /etc/systemd/system/app-schedule.service
```

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

[Install]
WantedBy=multi-user.target
```

```bash
# 启动服务
sudo systemctl enable app-schedule
sudo systemctl start app-schedule
sudo systemctl status app-schedule
```

## 5. 调试技巧

### 查看日志

```bash
tail -f storage/logs/app.log
```

### 测试单个任务

在 `kernel.go` 中临时添加：

```go
k.Cron("*/5 * * * * *", MyTask) // 每5秒执行一次
```

### 检查 cron 表达式

使用在线工具：https://crontab.guru/

注意：我们的表达式是 6 位（包含秒），标准 cron 是 5 位。

## 常见问题

**Q: 调度器和 Web 服务器可以同时运行吗？**  
A: 可以！它们是独立的进程。

**Q: 如何停止调度器？**  
A: 按 `Ctrl+C` 或使用 `systemctl stop app-schedule`

**Q: 任务执行时间超过间隔怎么办？**  
A: 会并发执行。如需避免，在任务中添加锁机制。

**Q: 如何确保只运行一个调度器实例？**  
A: 使用 systemd 或 supervisor 管理，它们会确保单实例运行。
