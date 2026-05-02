package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fiber-starter/config"

	"github.com/hibiken/asynq"
)

// QueueService 队列服务接口
type QueueService interface {
	Enqueue(taskName string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error)
	EnqueueIn(taskName string, payload interface{}, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error)
	EnqueueAt(taskName string, payload interface{}, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error)
	RegisterHandler(taskName string, handler asynq.HandlerFunc) error
	InspectQueues() ([]QueueStatus, error)
	StartWorker() error
	RunWorker() error
	StopWorker() error
	Close() error
}

type queueService struct {
	server     *asynq.Server
	client     *asynq.Client
	mux        *asynq.ServeMux
	config     *config.Config
	isRunning  bool
	mu         sync.Mutex
	clientOnce sync.Once
	serverOnce sync.Once
}

func NewQueueService(cfg *config.Config) QueueService {
	return &queueService{mux: asynq.NewServeMux(), config: cfg}
}

func (q *queueService) getRedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", q.config.Redis.Host, q.config.Redis.Port),
		Password: q.config.Redis.Password,
		DB:       q.config.Redis.DB + 1,
	}
}

func (q *queueService) getClient() *asynq.Client {
	q.clientOnce.Do(func() {
		q.client = asynq.NewClient(q.getRedisOpt())
	})
	return q.client
}

func (q *queueService) getServer() *asynq.Server {
	q.serverOnce.Do(func() {
		q.server = asynq.NewServer(
			q.getRedisOpt(),
			asynq.Config{
				Concurrency: q.config.Queue.Concurrency,
				RetryDelayFunc: func(n int, err error, task *asynq.Task) time.Duration {
					_ = err
					_ = task
					if n < 1 {
						n = 1
					}
					return time.Duration(n) * DefaultQueueRetryDelay
				},
				Queues: map[string]int{
					"critical": 6,
					"default":  3,
					"low":      1,
				},
			},
		)
	})
	return q.server
}

func (q *queueService) enqueue(taskName string, payload interface{}, enqueue func(*asynq.Client, *asynq.Task) (*asynq.TaskInfo, error), opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize task payload: %w", err)
	}

	task := asynq.NewTask(taskName, payloadBytes, opts...)
	info, err := enqueue(q.getClient(), task)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (q *queueService) Enqueue(taskName string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return q.enqueue(taskName, payload, func(client *asynq.Client, task *asynq.Task) (*asynq.TaskInfo, error) {
		info, err := client.Enqueue(task, q.defaultTaskOptions(opts...)...)
		if err != nil {
			return nil, fmt.Errorf("failed to add task to queue: %w", err)
		}
		return info, nil
	}, opts...)
}

func (q *queueService) EnqueueIn(taskName string, payload interface{}, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return q.enqueue(taskName, payload, func(client *asynq.Client, task *asynq.Task) (*asynq.TaskInfo, error) {
		allOpts := q.defaultTaskOptions(opts...)
		allOpts = append(allOpts, asynq.ProcessIn(delay))
		info, err := client.Enqueue(task, allOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to add task to queue with delay: %w", err)
		}
		return info, nil
	}, opts...)
}

func (q *queueService) EnqueueAt(taskName string, payload interface{}, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return q.enqueue(taskName, payload, func(client *asynq.Client, task *asynq.Task) (*asynq.TaskInfo, error) {
		allOpts := q.defaultTaskOptions(opts...)
		allOpts = append(allOpts, asynq.ProcessAt(at))
		info, err := client.Enqueue(task, allOpts...)
		if err != nil {
			return nil, fmt.Errorf("failed to add task to queue at specified time: %w", err)
		}
		return info, nil
	}, opts...)
}

func (q *queueService) RegisterHandler(taskName string, handler asynq.HandlerFunc) error {
	q.mux.HandleFunc(taskName, handler)
	return nil
}

func (q *queueService) InspectQueues() ([]QueueStatus, error) {
	inspector := asynq.NewInspector(q.getRedisOpt())
	defer func() { _ = inspector.Close() }()

	names, err := inspector.Queues()
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}

	statuses := make([]QueueStatus, 0, len(names))
	for _, name := range names {
		info, err := inspector.GetQueueInfo(name)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect queue %s: %w", name, err)
		}
		statuses = append(statuses, QueueStatus{
			Name:        info.Queue,
			Pending:     info.Pending,
			Running:     info.Active,
			Succeeded:   info.Completed,
			Failed:      info.Failed,
			Scheduled:   info.Scheduled,
			Retry:       info.Retry,
			Archived:    info.Archived,
			Paused:      info.Paused,
			Size:        info.Size,
			Processed:   info.Processed,
			TotalFail:   info.FailedTotal,
			MemoryUsage: info.MemoryUsage,
		})
	}

	return statuses, nil
}

func (q *queueService) defaultTaskOptions(opts ...asynq.Option) []asynq.Option {
	all := make([]asynq.Option, 0, len(opts)+1)
	all = append(all, asynq.MaxRetry(DefaultQueueRetryCount))
	all = append(all, opts...)
	return all
}
