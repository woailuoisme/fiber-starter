package providers_test

import (
	"testing"

	"lfiber/configs"
	mail "lfiber/internal/providers/mail"
	mailDrivers "lfiber/internal/providers/mail/drivers"
	queue "lfiber/internal/providers/queue"
	queueDrivers "lfiber/internal/providers/queue/drivers"
	storage "lfiber/internal/providers/storage"
	storageDrivers "lfiber/internal/providers/storage/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNoopFallback_Storage 验证当请求不支持的驱动或发生配置错误时，存储管理器能安全降级到 NoopDisk，避免崩溃。
func TestNoopFallback_Storage(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Storage.Driver = "invalid_driver"

	manager := storage.NewManager(cfg)
	require.NotNil(t, manager)

	// 验证请求不支持的驱动不发生 panic 而是返回 NoopDisk
	disk := manager.Disk()
	require.NotNil(t, disk)
	assert.IsType(t, &storageDrivers.NoopDisk{}, disk)

	// 验证 NoopDisk 的核心操作安全返回，不崩溃
	err := disk.Put("test.txt", []byte("hello"))
	require.NoError(t, err)

	exists, err := disk.Exists("test.txt")
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = disk.Get("test.txt")
	assert.Error(t, err)
}

// TestNoopFallback_Mail 验证在配置文件禁用邮件发送时，系统安全注入 NoopMailer 代替 nil。
func TestNoopFallback_Mail(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Mail.Enabled = false

	manager, mailer, err := mail.Register(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
	require.NotNil(t, mailer)

	// 验证获取的 Mailer 驱动为 NoopMailer
	assert.IsType(t, &mailDrivers.NoopMailer{}, mailer)

	// 验证邮件发送调用不崩溃
	msg := mailer.To("receiver@test.com").Subject("test").Plain("body")
	err = mailer.Send(msg)
	require.NoError(t, err)

	err = mailer.Raw("receiver@test.com", "test", "body")
	require.NoError(t, err)
}

// TestNoopFallback_Queue 验证在配置文件禁用异步队列时，系统安全注入 NoopQueue 代替 nil。
func TestNoopFallback_Queue(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Queue.Enabled = false

	manager, queueService, err := queue.RegisterQueue(cfg)
	require.NoError(t, err)
	require.NotNil(t, manager)
	require.NotNil(t, queueService)

	// 验证获取的 Queue 驱动为 NoopQueue
	assert.IsType(t, &queueDrivers.NoopQueue{}, queueService)

	// 验证任务投递不崩溃
	err = queueService.Push(nil)
	require.NoError(t, err)
}
