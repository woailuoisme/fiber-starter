package providers

import (
	"testing"

	mailContracts "lfiber/internal/providers/mail/contracts"
	notification "lfiber/internal/providers/notification"
	notificationContracts "lfiber/internal/providers/notification/contracts"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMessage is a mock of the Message interface.
type MockMessage struct {
	mock.Mock
}

func (m *MockMessage) To(to ...string) mailContracts.Message        { m.Called(to); return m }
func (m *MockMessage) Cc(cc ...string) mailContracts.Message        { m.Called(cc); return m }
func (m *MockMessage) Bcc(bcc ...string) mailContracts.Message      { m.Called(bcc); return m }
func (m *MockMessage) Subject(subject string) mailContracts.Message { m.Called(subject); return m }
func (m *MockMessage) Html(body string) mailContracts.Message       { m.Called(body); return m }
func (m *MockMessage) Plain(body string) mailContracts.Message      { m.Called(body); return m }
func (m *MockMessage) Attach(filePath string) mailContracts.Message { m.Called(filePath); return m }
func (m *MockMessage) Data(data map[string]interface{}) mailContracts.Message {
	m.Called(data)
	return m
}

func (m *MockMessage) View(templateName string, data map[string]interface{}) mailContracts.Message {
	m.Called(templateName, data)
	return m
}

func (m *MockMessage) Mailable(ml mailContracts.Mailable) mailContracts.Message {
	m.Called(ml)
	return m
}
func (m *MockMessage) GetTo() []string                 { return nil }
func (m *MockMessage) GetCc() []string                 { return nil }
func (m *MockMessage) GetBcc() []string                { return nil }
func (m *MockMessage) GetSubject() string              { return "" }
func (m *MockMessage) GetBody() string                 { return "" }
func (m *MockMessage) IsHtml() bool                    { return false }
func (m *MockMessage) GetAttachments() []string        { return nil }
func (m *MockMessage) GetData() map[string]interface{} { return nil }

// MockMailer is a mock of the Mailer interface.
type MockMailer struct {
	mock.Mock
}

func (m *MockMailer) To(to ...string) mailContracts.Message {
	args := m.Called(to)
	return args.Get(0).(mailContracts.Message)
}

func (m *MockMailer) Send(message mailContracts.Message) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockMailer) Raw(to, subject, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

func (m *MockMailer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMailer) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

// TestNotification implements the Notification contracts.
type TestNotification struct{}

func (n *TestNotification) Via(notifiable interface{}) []string {
	return []string{"mail"}
}

func (n *TestNotification) ToMail(notifiable interface{}) interface{} {
	return "Welcome to our system!"
}

// TestUser implements RouteNotificationForMail.
type TestUser struct {
	Email string
}

func (u *TestUser) RouteNotificationForMail() string {
	return u.Email
}

func TestNotificationProvider(t *testing.T) {
	mockMailer := new(MockMailer)
	mockMessage := new(MockMessage)
	manager := notification.NewNotificationManager(mockMailer)

	user := &TestUser{Email: "test@example.com"}
	notif := &TestNotification{}

	// Expectation
	mockMailer.On("To", []string{"test@example.com"}).Return(mockMessage)
	mockMailer.On("Send", mockMessage).Return(nil)

	err := manager.Send(user, notif)

	require.NoError(t, err)
	mockMailer.AssertExpectations(t)
}

func TestNotificationProvider_Extend(t *testing.T) {
	mockMailer := new(MockMailer)
	manager := notification.NewNotificationManager(mockMailer)
	channel := new(MockChannel)

	manager.Extend("sms", channel)

	notif := &TestSMSNotification{}
	user := &TestUser{Email: "test@example.com"}

	require.NoError(t, manager.Send(user, notif))
	assert.True(t, channel.Called)
}

type MockChannel struct {
	Called bool
}

func (c *MockChannel) Send(notifiable interface{}, notification notificationContracts.Notification) error {
	c.Called = true
	return nil
}

type TestSMSNotification struct{}

func (n *TestSMSNotification) Via(notifiable interface{}) []string {
	return []string{"sms"}
}
