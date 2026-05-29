package medialibrary

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	queueContracts "lfiber/internal/providers/queue/contracts"
)

const conversionJobTaskName = "medialibrary:perform-conversions"

var (
	defaultServiceMu sync.RWMutex
	defaultService   *Service
)

type ConversionJob struct {
	MediaID int64             `json:"media_id"`
	Rules   []*ConversionRule `json:"rules"`
	Queue   string            `json:"queue"`
}

func (j ConversionJob) Handle(ctx context.Context) error {
	service := currentService()
	if service == nil {
		return errors.New("media library service is not registered")
	}
	return service.RunConversions(ctx, j.MediaID, j.Rules)
}

func (j ConversionJob) TaskName() string {
	return conversionJobTaskName
}

func (j ConversionJob) QueueName() string {
	if j.Queue == "" {
		return "default"
	}
	return j.Queue
}

func RegisterJobs(reg interface{ Job(queueContracts.Job) }) {
	reg.Job(ConversionJob{})
}

func enqueueConversionJob(queue queueContracts.Queue, mediaID int64, rules []*ConversionRule, queueName string) error {
	if queue == nil {
		return errors.New("media library queue is not configured")
	}
	job := ConversionJob{MediaID: mediaID, Rules: rules, Queue: queueName}
	if queueName != "" {
		return queue.PushOn(queueName, job)
	}
	return queue.Push(job)
}

func setDefaultService(service *Service) {
	defaultServiceMu.Lock()
	defer defaultServiceMu.Unlock()
	defaultService = service
}

func currentService() *Service {
	defaultServiceMu.RLock()
	defer defaultServiceMu.RUnlock()
	return defaultService
}

func now() time.Time {
	return time.Now()
}

func wrapJobEnqueueError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("enqueue media conversion job: %w", err)
}
