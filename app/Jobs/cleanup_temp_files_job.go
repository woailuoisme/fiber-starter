package Jobs

import (
	"context"
	"fmt"
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
	fmt.Printf("Cleaning up temporary files in %s\n", j.Directory)
	return nil
}
