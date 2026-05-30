package providers_test

import (
	"testing"

	"lfiber/configs"
	providers "lfiber/internal/providers"
	mail "lfiber/internal/providers/mail"
	"lfiber/internal/providers/mail/drivers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMailTemplate_Rendering(t *testing.T) {
	// 临时设置全局容器，用来测试配置的默认值 fallback
	// Set global container runtime temporarily to test config variables fallback
	cfg := &configs.Config{}
	cfg.App.Name = "Test Fiber App"
	cfg.App.URL = "http://test-url.com"
	cfg.Mail.Enabled = true

	manager := mail.NewManager(cfg)
	runtime := &providers.Runtime{
		Config:       cfg,
		MailManager:  manager,
		EmailService: manager.Drive("log"),
	}
	providers.SetInstance(runtime)
	defer func() {
		_ = runtime.Close()
	}()

	t.Run("Render verification code template", func(t *testing.T) {
		// 使用 BaseMessage 类型的 View 方法进行单元渲染测试
		// Test rendering OTP verification code via BaseMessage.View
		msg := mail.Drive("log").To("recipient@example.com").View("verification_code", map[string]interface{}{
			"Code":      "123456",
			"ExpiresIn": "5 minutes",
			"Name":      "John Doe",
		})

		body := msg.GetBody()
		assert.Contains(t, body, "123456")
		assert.Contains(t, body, "5 minutes")
		assert.Contains(t, body, "John Doe")
		assert.Contains(t, body, "Test Fiber App") // 从配置中读取的 AppName
		assert.Contains(t, body, "http://test-url.com")
	})

	t.Run("Render reset password template", func(t *testing.T) {
		// 使用 mail.Message 类型的 View 方法进行渲染测试，并检查按钮样式
		// Test password reset email rendering via mail.Message.View
		msg := mail.NewMessage().View("reset_password", map[string]interface{}{
			"Name":      "Jane Doe",
			"ExpiresIn": "30 minutes",
			"Button": map[string]interface{}{
				"Color": "success",
				"Url":   "http://reset.com/123",
				"Text":  "Reset Button",
			},
		})

		body := msg.GetBody()
		assert.Contains(t, body, "Jane Doe")
		assert.Contains(t, body, "30 minutes")
		assert.Contains(t, body, "http://reset.com/123")
		assert.Contains(t, body, "Reset Button")
		assert.Contains(t, body, "#10b981") // 绿色的十六进制颜色码
	})

	t.Run("Render welcome template", func(t *testing.T) {
		// 测试欢迎激活邮件的渲染
		// Test welcome onboarding template rendering
		msg := mail.NewMessage().View("welcome", map[string]interface{}{
			"Name": "Bob Smith",
			"Button": map[string]interface{}{
				"Color": "primary",
				"Url":   "http://welcome.com",
				"Text":  "Get Started",
			},
		})

		body := msg.GetBody()
		assert.Contains(t, body, "Welcome to Test Fiber App")
		assert.Contains(t, body, "Bob Smith")
		assert.Contains(t, body, "Get Started")
		assert.Contains(t, body, "#4f46e5") // 蓝紫色的十六进制颜色码
	})

	t.Run("Render alert template", func(t *testing.T) {
		// 测试带详情表格的系统告警模板渲染
		// Test alert template rendering with details table
		details := map[string]interface{}{
			"IP Address": "192.168.1.1",
			"User Agent": "Mozilla/5.0",
		}
		msg := mail.NewMessage().View("alert", map[string]interface{}{
			"Title": "Suspicious Login",
			"Panel": map[string]interface{}{
				"Color": "danger",
				"Text":  "A login attempt was detected from a new device.",
			},
			"Details": details,
		})

		body := msg.GetBody()
		assert.Contains(t, body, "Suspicious Login")
		assert.Contains(t, body, "A login attempt was detected from a new device.")
		assert.Contains(t, body, "IP Address")
		assert.Contains(t, body, "192.168.1.1")
		assert.Contains(t, body, "User Agent")
		assert.Contains(t, body, "Mozilla/5.0")
		assert.Contains(t, body, "#f43f5e") // 红色的十六进制颜色码 (border)
	})

	t.Run("Fallback to default values if context is missing", func(t *testing.T) {
		// 清理全局 Runtime，测试在未挂载应用容器时的默认值 fallback
		// Clear global instance to test defaults when runtime is missing
		providers.SetInstance(nil)

		html, err := drivers.RenderTemplate("verification_code", map[string]interface{}{
			"Code": "9999",
		})
		require.NoError(t, err)
		assert.Contains(t, html, "Go Fiber App")          // 默认的 APP_NAME
		assert.Contains(t, html, "http://localhost:3300") // 默认的 APP_URL
		assert.Contains(t, html, "9999")
	})
}
