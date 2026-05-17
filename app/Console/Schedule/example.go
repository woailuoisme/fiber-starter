package schedule_tasks

import (
	"fiber-starter/app/Jobs"
	schedule "fiber-starter/app/Providers/Schedule"
)

// RegisterSampleTasks 展示了如何在不同的频率下注册调度任务
func RegisterSampleTasks() {
	// 1. 每隔五分钟清理一次临时文件
	schedule.Job(Jobs.NewCleanupTempFilesJob("/tmp/exports")).EveryFiveMinutes()

	// 2. 每天凌晨 3:00 发送汇总提醒
	schedule.Job(Jobs.NewReminderNotificationJob(1, "Daily Report Ready")).DailyAt("03:00")

	// 3. 每周一上午 9:00 执行
	schedule.Job(Jobs.NewReminderNotificationJob(1, "Weekly Planning")).WeeklyOn(1, "09:00")
}
