package schedule

import (
	"fiber-starter/app/helpers"
	"go.uber.org/zap"
)

// ExampleTask 示例任务
func ExampleTask() {
	helpers.Info("执行示例任务")
	// 在这里编写你的任务逻辑
}

// CleanupTask 清理任务示例
func CleanupTask() {
	helpers.Info("执行清理任务")
	// 清理临时文件、过期数据等
}

// SendEmailTask 发送邮件任务示例
func SendEmailTask() {
	helpers.Info("执行发送邮件任务")
	// 发送定时邮件
}

// BackupDatabaseTask 数据库备份任务示例
func BackupDatabaseTask() {
	helpers.Info("执行数据库备份任务")
	// 备份数据库
}

// GenerateReportTask 生成报表任务示例
func GenerateReportTask() {
	helpers.Info("执行生成报表任务", zap.String("type", "daily"))
	// 生成日报、周报等
}
