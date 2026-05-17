package Drivers

import (
	"fiber-starter/app/Providers/Mail/Contracts"
)

type BaseMessage struct {
	ToArr       []string
	CcArr       []string
	BccArr      []string
	SubjectStr  string
	BodyStr     string
	IsHtmlFlag  bool
	Attachments []string
	DataMap     map[string]interface{}
}

func (m *BaseMessage) To(to ...string) Contracts.Message {
	m.ToArr = append(m.ToArr, to...)
	return m
}

func (m *BaseMessage) Cc(cc ...string) Contracts.Message {
	m.CcArr = append(m.CcArr, cc...)
	return m
}

func (m *BaseMessage) Bcc(bcc ...string) Contracts.Message {
	m.BccArr = append(m.BccArr, bcc...)
	return m
}

func (m *BaseMessage) Subject(subject string) Contracts.Message {
	m.SubjectStr = subject
	return m
}

func (m *BaseMessage) Html(body string) Contracts.Message {
	m.BodyStr = body
	m.IsHtmlFlag = true
	return m
}

func (m *BaseMessage) Plain(body string) Contracts.Message {
	m.BodyStr = body
	m.IsHtmlFlag = false
	return m
}

func (m *BaseMessage) Attach(filePath string) Contracts.Message {
	m.Attachments = append(m.Attachments, filePath)
	return m
}

func (m *BaseMessage) Data(data map[string]interface{}) Contracts.Message {
	m.DataMap = data
	return m
}

func (m *BaseMessage) GetTo() []string                 { return m.ToArr }
func (m *BaseMessage) GetCc() []string                 { return m.CcArr }
func (m *BaseMessage) GetBcc() []string                { return m.BccArr }
func (m *BaseMessage) GetSubject() string              { return m.SubjectStr }
func (m *BaseMessage) GetBody() string                 { return m.BodyStr }
func (m *BaseMessage) IsHtml() bool                    { return m.IsHtmlFlag }
func (m *BaseMessage) GetAttachments() []string        { return m.Attachments }
func (m *BaseMessage) GetData() map[string]interface{} { return m.DataMap }
