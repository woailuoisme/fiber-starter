package contracts

// Mailer defines the interface for sending emails (similar to Laravel's Mail)
type Mailer interface {
	// To creates a new message with the given recipient
	To(to ...string) Message

	// Send sends a message immediately
	Send(message Message) error

	// Raw sends a raw email
	Raw(to, subject, body string) error

	// HealthCheck verifies the mailer connection is alive
	HealthCheck() error

	// Close closes the mailer connection
	Close() error
}

// Mailable 接口定义了可邮寄邮件的属性和模板映射。
// 使用此接口能够实现强类型邮件模板设计，避免在业务代码中手写 untyped map 从而减少出错可能。
type Mailable interface {
	Subject() string
	Template() (name string, data map[string]interface{})
}

// Message defines the interface for building an email message
type Message interface {
	To(to ...string) Message
	Cc(cc ...string) Message
	Bcc(bcc ...string) Message
	Subject(subject string) Message
	Html(body string) Message
	Plain(body string) Message
	Attach(filePath string) Message
	Data(data map[string]interface{}) Message
	View(templateName string, data map[string]interface{}) Message
	Mailable(m Mailable) Message
	GetTo() []string
	GetCc() []string
	GetBcc() []string
	GetSubject() string
	GetBody() string
	IsHtml() bool
	GetAttachments() []string
	GetData() map[string]interface{}
}
