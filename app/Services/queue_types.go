package services

import "time"

const (
	TaskSendWelcomeEmail     = "send_welcome_email"
	TaskSendPasswordReset    = "send_password_reset_email"
	TaskSendVerification     = "send_verification_email"
	TaskProcessDataExport    = "process_data_export"
	TaskGenerateReport       = "generate_report"
	TaskCleanupTempFiles     = "cleanup_temp_files"
	TaskUpdateUserStatistics = "update_user_statistics"

	DefaultQueueName        = "default"
	DefaultQueueRetryCount  = 3
	DefaultQueueRetryDelay  = 30 * time.Second
	DefaultQueueTaskTimeout = 30 * time.Second
)

type (
	SendEmailPayload struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		IsHTML  bool   `json:"is_html"`
	}
	ProcessDataExportPayload struct {
		UserID     uint   `json:"user_id"`
		ExportType string `json:"export_type"`
		Format     string `json:"format"`
		Email      string `json:"email"`
	}
	GenerateReportPayload struct {
		ReportType string    `json:"report_type"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
		UserID     uint      `json:"user_id"`
		Email      string    `json:"email"`
	}
	UpdateUserStatisticsPayload struct {
		UserID uint `json:"user_id"`
	}
)
