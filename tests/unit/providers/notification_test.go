package providers

import (
	"testing"

	notification "lfiber/internal/providers/notification"
	notificationContracts "lfiber/internal/providers/notification/contracts"
	mailmocks "lfiber/tests/mocks/providers/mail"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	mockMailer := mailmocks.NewMockMailer(t)
	mockMessage := mailmocks.NewMockMessage(t)
	manager := notification.NewNotificationManager(mockMailer)

	user := &TestUser{Email: "test@example.com"}
	notif := &TestNotification{}

	// Expectation
	mockMailer.On("To", []string{"test@example.com"}).Return(mockMessage)
	mockMailer.On("Send", mockMessage).Return(nil)

	err := manager.Send(user, notif)

	require.NoError(t, err)
}

func TestNotificationProvider_Extend(t *testing.T) {
	mockMailer := mailmocks.NewMockMailer(t)
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
