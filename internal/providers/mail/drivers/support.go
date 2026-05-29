package drivers

import (
	"fmt"

	"lfiber/internal/providers/mail/contracts"
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

func (m *BaseMessage) To(to ...string) contracts.Message {
	m.ToArr = append(m.ToArr, to...)
	return m
}

func (m *BaseMessage) Cc(cc ...string) contracts.Message {
	m.CcArr = append(m.CcArr, cc...)
	return m
}

func (m *BaseMessage) Bcc(bcc ...string) contracts.Message {
	m.BccArr = append(m.BccArr, bcc...)
	return m
}

func (m *BaseMessage) Subject(subject string) contracts.Message {
	m.SubjectStr = subject
	return m
}

func (m *BaseMessage) Html(body string) contracts.Message {
	m.BodyStr = body
	m.IsHtmlFlag = true
	return m
}

func (m *BaseMessage) Plain(body string) contracts.Message {
	m.BodyStr = body
	m.IsHtmlFlag = false
	return m
}

func (m *BaseMessage) Attach(filePath string) contracts.Message {
	m.Attachments = append(m.Attachments, filePath)
	return m
}

func (m *BaseMessage) Data(data map[string]interface{}) contracts.Message {
	m.DataMap = data
	return m
}

func (m *BaseMessage) View(templateName string, data map[string]interface{}) contracts.Message {
	m.DataMap = data
	htmlContent, err := RenderTemplate(templateName, data)
	if err != nil {
		m.BodyStr = fmt.Sprintf("<!-- Error rendering template: %v -->", err)
	} else {
		m.BodyStr = htmlContent
	}
	m.IsHtmlFlag = true
	return m
}

// Mailable 根据传入的 Mailable 接口实例配置并构建邮件。
// 这样业务逻辑层就可以直接传入具有强类型的结构体以提升可读性和类型安全性。
func (m *BaseMessage) Mailable(ml contracts.Mailable) contracts.Message {
	name, data := ml.Template()
	m.View(name, data)
	m.Subject(ml.Subject())
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
