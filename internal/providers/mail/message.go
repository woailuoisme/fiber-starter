package mail

import (
	"lfiber/internal/providers/mail/contracts"
)

type Message struct {
	to          []string
	cc          []string
	bcc         []string
	subject     string
	body        string
	isHtml      bool
	attachments []string
	data        map[string]interface{}
}

func NewMessage() *Message {
	return &Message{}
}

func (m *Message) To(to ...string) contracts.Message {
	m.to = append(m.to, to...)
	return m
}

func (m *Message) Cc(cc ...string) contracts.Message {
	m.cc = append(m.cc, cc...)
	return m
}

func (m *Message) Bcc(bcc ...string) contracts.Message {
	m.bcc = append(m.bcc, bcc...)
	return m
}

func (m *Message) Subject(subject string) contracts.Message {
	m.subject = subject
	return m
}

func (m *Message) Html(body string) contracts.Message {
	m.body = body
	m.isHtml = true
	return m
}

func (m *Message) Plain(body string) contracts.Message {
	m.body = body
	m.isHtml = false
	return m
}

func (m *Message) Attach(filePath string) contracts.Message {
	m.attachments = append(m.attachments, filePath)
	return m
}

func (m *Message) Data(data map[string]interface{}) contracts.Message {
	m.data = data
	return m
}

func (m *Message) GetTo() []string                 { return m.to }
func (m *Message) GetCc() []string                 { return m.cc }
func (m *Message) GetBcc() []string                { return m.bcc }
func (m *Message) GetSubject() string              { return m.subject }
func (m *Message) GetBody() string                 { return m.body }
func (m *Message) IsHtml() bool                    { return m.isHtml }
func (m *Message) GetAttachments() []string        { return m.attachments }
func (m *Message) GetData() map[string]interface{} { return m.data }
