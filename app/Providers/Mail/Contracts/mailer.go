package Contracts

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
	GetTo() []string
	GetCc() []string
	GetBcc() []string
	GetSubject() string
	GetBody() string
	IsHtml() bool
	GetAttachments() []string
	GetData() map[string]interface{}
}
