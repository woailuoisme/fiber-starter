package jobs

import (
	"context"

	queueContracts "lfiber/internal/providers/queue/contracts"
	helpers "lfiber/internal/support"

	"go.uber.org/zap"
)

type CleanupTempFilesJob struct {
	queueContracts.JobMeta
	Directory string `json:"directory"`
}

func NewCleanupTempFilesJob(dir string) *CleanupTempFilesJob {
	return &CleanupTempFilesJob{
		JobMeta:   queueContracts.NewJobMeta("cleanup_temp_files", "low"),
		Directory: dir,
	}
}

func (j *CleanupTempFilesJob) Handle(ctx context.Context) error {
	// Logic to cleanup temp files
	helpers.Info("cleanup_temp_files_started", zap.String("directory", j.Directory))
	return nil
}
