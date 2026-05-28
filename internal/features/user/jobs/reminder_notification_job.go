package jobs

import (
	"context"

	queueContracts "lfiber/internal/providers/queue/contracts"
	helpers "lfiber/internal/support"

	"go.uber.org/zap"
)

// ReminderNotificationJob 模拟一个提醒通知任务
type ReminderNotificationJob struct {
	queueContracts.JobMeta
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func NewReminderNotificationJob(userID int64, message string) *ReminderNotificationJob {
	return &ReminderNotificationJob{
		JobMeta: queueContracts.NewJobMeta("reminder_notification", "default"),
		UserID:  userID,
		Message: message,
	}
}

func (j *ReminderNotificationJob) Handle(ctx context.Context) error {
	helpers.Info(
		"Processing reminder notification",
		zap.Int64("user_id", j.UserID),
		zap.String("message", j.Message),
	)
	// 实际业务逻辑：调用推送服务或发送短信
	return nil
}
