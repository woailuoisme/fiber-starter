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
	StartWorker() error
	StopWorker() error
}

// queueService 队列服务实现
type queueService struct {
	server     *asynq.Server
	client     *asynq.Client
	mux        *asynq.ServeMux
	config     *config.Config
	isRunning  bool
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

// Enqueue 添加任务到队列
func (q *queueService) Enqueue(taskName string, payload interface{}, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化任务负载失败: %w", err)
	}

	// 创建任务
	task := asynq.NewTask(taskName, payloadBytes, opts...)

	// 添加到队列
	info, err := q.getClient().Enqueue(task)
	if err != nil {
		return nil, fmt.Errorf("添加任务到队列失败: %w", err)
	}

	return info, nil
}

// EnqueueIn 延迟添加任务到队列
func (q *queueService) EnqueueIn(taskName string, payload interface{},
	delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化任务负载失败: %w", err)
	}

	// 创建任务
	task := asynq.NewTask(taskName, payloadBytes, opts...)

	// 延迟添加到队列
	info, err := q.getClient().Enqueue(task, asynq.ProcessIn(delay))
	if err != nil {
		return nil, fmt.Errorf("延迟添加任务到队列失败: %w", err)
	}

	return info, nil
}

// EnqueueAt 在指定时间添加任务到队列
func (q *queueService) EnqueueAt(taskName string, payload interface{},
	at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 序列化负载
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("序列化任务负载失败: %w", err)
	}

	// 创建任务
	task := asynq.NewTask(taskName, payloadBytes, opts...)

	// 在指定时间添加到队列
	info, err := q.getClient().Enqueue(task, asynq.ProcessAt(at))
	if err != nil {
		return nil, fmt.Errorf("定时添加任务到队列失败: %w", err)
	}

	return info, nil
}

// RegisterHandler 注册任务处理器
func (q *queueService) RegisterHandler(taskName string, handler asynq.HandlerFunc) error {
	q.mux.HandleFunc(taskName, handler)
	return nil
}

// StartWorker 启动工作进程
func (q *queueService) StartWorker() error {
	if q.isRunning {
		return fmt.Errorf("工作进程已在运行")
	}

	// 启动工作进程
	go func() {
		q.isRunning = true
		if err := q.getServer().Run(q.mux); err != nil {
			fmt.Printf("队列工作进程启动失败: %v\n", err)
		}
	}()

	return nil
}

// StopWorker 停止工作进程
func (q *queueService) StopWorker() error {
	if !q.isRunning {
		return fmt.Errorf("工作进程未在运行")
	}

	// 优雅关闭
	q.getServer().Shutdown()
	q.isRunning = false

	// 关闭客户端
	_ = q.getClient().Close()

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
