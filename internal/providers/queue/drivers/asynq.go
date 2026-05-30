package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"lfiber/configs"
	"lfiber/internal/providers/asynqredis"
	"lfiber/internal/providers/queue/contracts"
	helpers "lfiber/internal/support"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type AsynqDriver struct {
	server     *asynq.Server
	client     *asynq.Client
	mux        *asynq.ServeMux
	config     *configs.Config
	isRunning  bool
	mu         sync.Mutex
	clientOnce sync.Once
	serverOnce sync.Once

	concurrencyOverride int
}

func (d *AsynqDriver) SetConcurrency(n int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.concurrencyOverride = n
}

func (d *AsynqDriver) GetConcurrency() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.concurrencyOverride > 0 {
		return d.concurrencyOverride
	}
	return d.config.Queue.Concurrency
}

func NewAsynqDriver(cfg *configs.Config) *AsynqDriver {
	return &AsynqDriver{
		mux:    asynq.NewServeMux(),
		config: cfg,
	}
}

func (d *AsynqDriver) getRedisOpt() asynq.RedisConnOpt {
	return asynqredis.NewClientOpt(d.config, 1)
}

func (d *AsynqDriver) getClient() *asynq.Client {
	d.clientOnce.Do(func() {
		d.client = asynq.NewClient(d.getRedisOpt())
	})
	return d.client
}

func (d *AsynqDriver) getServer(queues ...string) *asynq.Server {
	d.serverOnce.Do(func() {
		queueMap := map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		}

		// If specific queues are provided, we filter or override the map
		if len(queues) > 0 {
			newMap := make(map[string]int)
			for _, q := range queues {
				if priority, ok := queueMap[q]; ok {
					newMap[q] = priority
				} else {
					newMap[q] = 1 // Default priority for unknown queues
				}
			}
			queueMap = newMap
		}

		d.server = asynq.NewServer(
			d.getRedisOpt(),
			asynq.Config{
				Concurrency: d.GetConcurrency(),
				Queues:      queueMap,
			},
		)
	})
	return d.server
}

func (d *AsynqDriver) Push(job contracts.Job) error {
	return d.enqueue(job, nil, "")
}

func (d *AsynqDriver) Size(queue ...string) (int64, error) {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	target := "default"
	if len(queue) > 0 && queue[0] != "" {
		target = queue[0]
	}

	info, err := inspector.GetQueueInfo(target)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue size: %w", err)
	}

	return int64(info.Pending), nil
}

func (d *AsynqDriver) PushOn(queue string, job contracts.Job) error {
	return d.enqueue(job, nil, queue)
}

func (d *AsynqDriver) Later(delay time.Duration, job contracts.Job) error {
	return d.enqueue(job, asynq.ProcessIn(delay), "")
}

func (d *AsynqDriver) LaterOn(queue string, delay time.Duration, job contracts.Job) error {
	return d.enqueue(job, asynq.ProcessIn(delay), queue)
}

func (d *AsynqDriver) Bulk(jobs []contracts.Job, queue ...string) error {
	q := ""
	if len(queue) > 0 {
		q = queue[0]
	}

	for _, job := range jobs {
		if err := d.enqueue(job, nil, q); err != nil {
			return err
		}
	}
	return nil
}

func (d *AsynqDriver) ProcessAt(at time.Time, job contracts.Job) error {
	return d.enqueue(job, asynq.ProcessAt(at), "")
}

func (d *AsynqDriver) enqueue(job contracts.Job, opt asynq.Option, queue string) error {
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job payload: %w", err)
	}

	task := asynq.NewTask(job.TaskName(), payload)
	var opts []asynq.Option

	// Add the queue name option
	if queue != "" {
		opts = append(opts, asynq.Queue(queue))
	} else {
		opts = append(opts, asynq.Queue(job.QueueName()))
	}

	if opt != nil {
		opts = append(opts, opt)
	}

	_, err = d.getClient().Enqueue(task, opts...)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

func (d *AsynqDriver) Register(job contracts.Job) {
	jobType := reflect.TypeOf(job)
	if jobType.Kind() == reflect.Pointer {
		jobType = jobType.Elem()
	}

	d.mux.HandleFunc(job.TaskName(), func(ctx context.Context, t *asynq.Task) error {
		// Create a new instance of the job
		newJob := reflect.New(jobType).Interface().(contracts.Job)

		if err := json.Unmarshal(t.Payload(), &newJob); err != nil {
			return fmt.Errorf("failed to unmarshal job payload: %w", err)
		}

		helpers.Info("Processing job", zap.String("task", job.TaskName()))
		return newJob.Handle(ctx)
	})
}

func (d *AsynqDriver) StartWorker(queues ...string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isRunning {
		return fmt.Errorf("worker is already running")
	}

	go func() {
		if err := d.getServer(queues...).Run(d.mux); err != nil {
			helpers.LogError("Queue worker failed", zap.Error(err))
		}
	}()

	d.isRunning = true
	helpers.Info("Queue worker started (non-blocking)", zap.Strings("queues", queues))
	return nil
}

func (d *AsynqDriver) RunWorker(queues ...string) error {
	d.mu.Lock()
	if d.isRunning {
		d.mu.Unlock()
		return fmt.Errorf("worker is already running")
	}
	d.isRunning = true
	d.mu.Unlock()

	helpers.Info("Queue worker starting (blocking)...", zap.Strings("queues", queues))
	err := d.getServer(queues...).Run(d.mux)

	d.mu.Lock()
	d.isRunning = false
	d.mu.Unlock()

	return err
}

func (d *AsynqDriver) StopWorker() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isRunning {
		return nil
	}

	if d.server != nil {
		d.server.Shutdown()
		d.server = nil
	}
	d.isRunning = false
	helpers.Info("Queue worker stopped")
	return nil
}

func (d *AsynqDriver) InspectQueues() ([]contracts.QueueStatus, error) {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	queues, err := inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queues: %w", err)
	}

	var statuses []contracts.QueueStatus
	for _, q := range queues {
		info, err := inspector.GetQueueInfo(q)
		if err != nil {
			continue
		}

		statuses = append(statuses, contracts.QueueStatus{
			Name:      q,
			Pending:   info.Pending,
			Running:   info.Active,
			Succeeded: info.Processed,
			Failed:    info.Failed,
			Scheduled: info.Scheduled,
			Retry:     info.Retry,
			Archived:  info.Archived,
			Paused:    info.Paused,
		})
	}

	return statuses, nil
}

func (d *AsynqDriver) ListFailed(page, pageSize int) ([]contracts.FailedJob, error) {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	// List from the 'default' queue by default for simplicity, or we'd need to iterate all queues
	tasks, err := inspector.ListArchivedTasks("default", asynq.Page(page), asynq.PageSize(pageSize))
	if err != nil {
		return nil, fmt.Errorf("failed to list failed tasks: %w", err)
	}

	var failedJobs []contracts.FailedJob
	for _, t := range tasks {
		failedJobs = append(failedJobs, contracts.FailedJob{
			ID:         t.ID,
			Payload:    string(t.Payload),
			Queue:      t.Queue,
			FailedAt:   t.LastFailedAt,
			Error:      t.LastErr,
			Retried:    t.Retried,
			MaxRetries: t.MaxRetry,
		})
	}

	return failedJobs, nil
}

func (d *AsynqDriver) RetryFailed(id string) error {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	// Moving from archived back to pending
	err := inspector.RunTask("default", id)
	if err != nil {
		return fmt.Errorf("failed to retry task %s: %w", id, err)
	}

	return nil
}

func (d *AsynqDriver) DeleteFailed(id string) error {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	err := inspector.DeleteTask("default", id)
	if err != nil {
		return fmt.Errorf("failed to delete task %s: %w", id, err)
	}

	return nil
}

func (d *AsynqDriver) Flush(queue string) error {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	_, err := inspector.DeleteAllArchivedTasks(queue)
	if err != nil {
		return fmt.Errorf("failed to flush archived tasks for queue %s: %w", queue, err)
	}

	return nil
}

func (d *AsynqDriver) HealthCheck() error {
	inspector := asynq.NewInspector(d.getRedisOpt())
	defer func() { _ = inspector.Close() }()
	_, err := inspector.Queues()
	return err
}

func (d *AsynqDriver) Close() error {
	_ = d.StopWorker()
	if d.client != nil {
		err := d.client.Close()
		d.client = nil
		return err
	}
	return nil
}
