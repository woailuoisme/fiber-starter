package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// QueueService 队列服务接口
type QueueService interface {
	Enqueue(taskName string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error)
	EnqueueIn(taskName string, payload interface{}, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error)
	EnqueueAt(taskName string, payload interface{}, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error)
	RegisterHandler(taskName string, handler asynq.HandlerFunc) error
	StartWorker() error
	RunWorker() error
	StopWorker() error
	Close() error
}

// queueService 队列服务实现
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

// NewQueueService 创建队列服务实例
func NewQueueService(cfg *config.Config) QueueService {
	// 创建路由器
	mux := asynq.NewServeMux()

	return &queueService{
		mux:    mux,
		config: cfg,
	}
}

// getRedisOpt 获取Redis配置
func (q *queueService) getRedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", q.config.Redis.Host, q.config.Redis.Port),
		Password: q.config.Redis.Password,
		DB:       q.config.Redis.DB + 1, // 使用不同的DB
	}
}

// getClient 获取或初始化客户端（懒加载）
func (q *queueService) getClient() *asynq.Client {
	q.clientOnce.Do(func() {
		q.client = asynq.NewClient(q.getRedisOpt())
	})
	return q.client
}

// getServer 获取或初始化服务器（懒加载）
func (q *queueService) getServer() *asynq.Server {
	q.serverOnce.Do(func() {
		q.server = asynq.NewServer(
			q.getRedisOpt(),
			asynq.Config{
				Concurrency: q.config.Queue.Concurrency,
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

// Enqueue Add task to queue
func (q *queueService) Enqueue(taskName string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize task payload: %w", err)
	}

	// Create task
	task := asynq.NewTask(taskName, payloadBytes, opts...)

	// Add to queue
	info, err := q.getClient().Enqueue(task)
	if err != nil {
		return nil, fmt.Errorf("failed to add task to queue: %w", err)
	}

	return info, nil
}

// EnqueueIn Add task to queue with delay
func (q *queueService) EnqueueIn(taskName string, payload interface{},
	delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize task payload: %w", err)
	}

	// Create task
	task := asynq.NewTask(taskName, payloadBytes, opts...)

	// Delayed add to queue
	info, err := q.getClient().Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		return nil, fmt.Errorf("failed to add task to queue with delay: %w", err)
	}

	return info, nil
}

// EnqueueAt Add task to queue at specified time
func (q *queueService) EnqueueAt(taskName string, payload interface{},
	at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// Serialize payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize task payload: %w", err)
	}

	// Create task
	task := asynq.NewTask(taskName, payloadBytes, opts...)

	// Add to queue at specified time
	info, err := q.getClient().Enqueue(task, asynq.ProcessAt(at))
	if err != nil {
		return nil, fmt.Errorf("failed to add task to queue at specified time: %w", err)
	}

	return info, nil
}

// RegisterHandler Register task handler
func (q *queueService) RegisterHandler(taskName string, handler asynq.HandlerFunc) error {
	q.mux.HandleFunc(taskName, handler)
	return nil
}

// StartWorker Start worker process
func (q *queueService) StartWorker() error {
	q.mu.Lock()
	if q.isRunning {
		q.mu.Unlock()
		return fmt.Errorf("worker process is already running")
	}
	q.isRunning = true
	q.mu.Unlock()

	go func() {
		if err := q.getServer().Run(q.mux); err != nil {
			helpers.LogError("Queue worker process start failed", zap.Error(err))
		}
		q.mu.Lock()
		q.isRunning = false
		q.mu.Unlock()
	}()

	return nil
}

func (q *queueService) RunWorker() error {
	q.mu.Lock()
	if q.isRunning {
		q.mu.Unlock()
		return fmt.Errorf("worker process is already running")
	}
	q.isRunning = true
	q.mu.Unlock()

	err := q.getServer().Run(q.mux)

	q.mu.Lock()
	q.isRunning = false
	q.mu.Unlock()

	return err
}

// StopWorker Stop worker process
func (q *queueService) StopWorker() error {
	q.mu.Lock()
	if !q.isRunning {
		q.mu.Unlock()
		return fmt.Errorf("worker process is not running")
	}
	q.mu.Unlock()

	// 优雅关闭
	q.getServer().Shutdown()
	q.mu.Lock()
	q.isRunning = false
	q.mu.Unlock()

	// 关闭客户端
	_ = q.getClient().Close()

	return nil
}

func (q *queueService) Close() error {
	if q.isRunning {
		q.getServer().Shutdown()
		q.isRunning = false
	}
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

// 预定义任务类型
const (
	TaskSendWelcomeEmail     = "send_welcome_email"
	TaskSendPasswordReset    = "send_password_reset_email"
	TaskSendVerification     = "send_verification_email"
	TaskProcessDataExport    = "process_data_export"
	TaskGenerateReport       = "generate_report"
	TaskCleanupTempFiles     = "cleanup_temp_files"
	TaskUpdateUserStatistics = "update_user_statistics"
)

// 任务负载结构
type (
	// SendEmailPayload 发送邮件负载
	SendEmailPayload struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
		IsHTML  bool   `json:"is_html"`
	}

	// ProcessDataExportPayload 处理数据导出负载
	ProcessDataExportPayload struct {
		UserID     uint   `json:"user_id"`
		ExportType string `json:"export_type"`
		Format     string `json:"format"`
		Email      string `json:"email"`
	}

	// GenerateReportPayload 生成报告负载
	GenerateReportPayload struct {
		ReportType string    `json:"report_type"`
		StartDate  time.Time `json:"start_date"`
		EndDate    time.Time `json:"end_date"`
		UserID     uint      `json:"user_id"`
		Email      string    `json:"email"`
	}

	// UpdateUserStatisticsPayload 更新用户统计负载
	UpdateUserStatisticsPayload struct {
		UserID uint `json:"user_id"`
	}
)
