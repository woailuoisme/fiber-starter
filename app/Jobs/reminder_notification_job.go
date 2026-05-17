package Jobs

import (
	"context"

	helpers "fiber-starter/app/Support"

	"go.uber.org/zap"
)

// ReminderNotificationJob 模拟一个提醒通知任务
type ReminderNotificationJob struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func NewReminderNotificationJob(userID int64, message string) *ReminderNotificationJob {
	return &ReminderNotificationJob{
		UserID:  userID,
		Message: message,
	}
}

func (j *ReminderNotificationJob) TaskName() string {
	return "reminder_notification"
}

func (j *ReminderNotificationJob) QueueName() string {
	return "default"
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
