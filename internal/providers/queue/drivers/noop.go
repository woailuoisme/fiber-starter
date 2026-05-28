package drivers

import (
	"time"

	"lfiber/internal/providers/queue/contracts"
)

// NoopQueue 实现 contracts.Queue 接口。
// 用于在异步队列禁用或降级时作为占位驱动，避免空指针和阻塞调用。
type NoopQueue struct {
	concurrency int
}

var _ contracts.Queue = (*NoopQueue)(nil)

// NewNoopQueue 创建并返回一个 NoopQueue 实例。
func NewNoopQueue() *NoopQueue {
	return &NoopQueue{concurrency: 1}
}

func (n *NoopQueue) Push(job contracts.Job) error {
	return nil
}

func (n *NoopQueue) Size(queue ...string) (int64, error) {
	return 0, nil
}

func (n *NoopQueue) PushOn(queue string, job contracts.Job) error {
	return nil
}

func (n *NoopQueue) Later(delay time.Duration, job contracts.Job) error {
	return nil
}

func (n *NoopQueue) LaterOn(queue string, delay time.Duration, job contracts.Job) error {
	return nil
}

func (n *NoopQueue) Bulk(jobs []contracts.Job, queue ...string) error {
	return nil
}

func (n *NoopQueue) ProcessAt(at time.Time, job contracts.Job) error {
	return nil
}

func (n *NoopQueue) Register(job contracts.Job) {}

func (n *NoopQueue) StartWorker(queue ...string) error {
	return nil
}

func (n *NoopQueue) RunWorker(queue ...string) error {
	return nil
}

func (n *NoopQueue) StopWorker() error {
	return nil
}

func (n *NoopQueue) InspectQueues() ([]contracts.QueueStatus, error) {
	return []contracts.QueueStatus{}, nil
}

func (n *NoopQueue) ListFailed(page, pageSize int) ([]contracts.FailedJob, error) {
	return []contracts.FailedJob{}, nil
}

func (n *NoopQueue) RetryFailed(id string) error {
	return nil
}

func (n *NoopQueue) DeleteFailed(id string) error {
	return nil
}

func (n *NoopQueue) Flush(queue string) error {
	return nil
}

func (n *NoopQueue) HealthCheck() error {
	return nil
}

func (n *NoopQueue) Close() error {
	return nil
}

func (n *NoopQueue) SetConcurrency(num int) {
	n.concurrency = num
}

func (n *NoopQueue) GetConcurrency() int {
	return n.concurrency
}
