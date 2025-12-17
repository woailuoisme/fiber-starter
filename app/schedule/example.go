package schedule

import (
	"time"

	"go.uber.org/zap"

	"fiber-starter/app/helpers"
)

// TestScheduleTask 测试定时任务（每10秒执行一次）
func TestScheduleTask() {
	now := time.Now().Format("2006-01-02 15:04:05")
	helpers.Info("🎯 定时任务测试", zap.String("time", now))
}

// 如果你想快速测试定时任务是否工作，可以在 kernel.go 的 Schedule() 方法中添加：
//
// // 测试：每10秒执行一次
// k.Cron("*/10 * * * * *", TestScheduleTask)
//
// 然后运行：make schedule
// 你会看到每10秒输出一次日志
