package drivers

import (
	"lfiber/internal/providers/mail/contracts"
)

// NoopMailer 实现 contracts.Mailer 接口。
// 用于在邮件服务禁用或发生故障降级时，作为安全的占位驱动，避免空指针崩溃。
type NoopMailer struct{}

var _ contracts.Mailer = (*NoopMailer)(nil)

// NewNoopMailer 创建并返回一个 NoopMailer 实例。
func NewNoopMailer() *NoopMailer {
	return &NoopMailer{}
}

func (n *NoopMailer) To(to ...string) contracts.Message {
	msg := &BaseMessage{}
	return msg.To(to...)
}

func (n *NoopMailer) Send(message contracts.Message) error {
	return nil
}

func (n *NoopMailer) Raw(to, subject, body string) error {
	return nil
}

func (n *NoopMailer) HealthCheck() error {
	return nil
}

func (n *NoopMailer) Close() error {
	return nil
}
