package jobs

import (
	"context"

	helpers "lfiber/internal/support"

	"go.uber.org/zap"
)

type CleanupTempFilesJob struct {
	Directory string `json:"directory"`
}

func NewCleanupTempFilesJob(dir string) *CleanupTempFilesJob {
	return &CleanupTempFilesJob{
		Directory: dir,
	}
}

func (j *CleanupTempFilesJob) TaskName() string {
	return "cleanup_temp_files"
}

func (j *CleanupTempFilesJob) QueueName() string {
	return "low"
}

func (j *CleanupTempFilesJob) Handle(ctx context.Context) error {
	// Logic to cleanup temp files
	helpers.Info("cleanup_temp_files_started", zap.String("directory", j.Directory))
	return nil
}
