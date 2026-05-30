package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"testing"

	"lfiber/configs"
	"lfiber/internal/bootstrap"
	"lfiber/internal/features/user"
	"lfiber/internal/providers"
	authProvider "lfiber/internal/providers/auth"
	authorizationProvider "lfiber/internal/providers/authorization"
	cacheProvider "lfiber/internal/providers/cache"
	dbProvider "lfiber/internal/providers/database"
	hashProvider "lfiber/internal/providers/hash"
	i18nProvider "lfiber/internal/providers/i18n"
	mailContracts "lfiber/internal/providers/mail/contracts"
	"lfiber/internal/providers/mail/drivers"
	"lfiber/tests/internal/testkit"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthEndpoints_SignUpAndLogin(t *testing.T) {
	app, mailer, _ := newTestApp(t)

	// 1. SignUp
	signUpReq := map[string]string{
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "password123",
	}
	body, err := json.Marshal(signUpReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var signUpResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			VerificationRequired bool `json:"verification_required"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signUpResult))
	assert.True(t, signUpResult.Success)
	assert.True(t, signUpResult.Data.VerificationRequired)

	// Extract verification code from captured mock emails
	require.Len(t, mailer.rawCalls, 1)
	code := extractOTPCode(t, mailer.rawCalls[0].Body)

	// 2. Verify SignUp
	verifyReq := map[string]string{
		"email": "bob@example.com",
		"code":  code,
	}
	body, err = json.Marshal(verifyReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/sign-up/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// 3. SignIn
	signInReq := map[string]string{
		"email":    "bob@example.com",
		"password": "password123",
	}
	body, err = json.Marshal(signInReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var signInResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			Tokens struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signInResult))
	assert.True(t, signInResult.Success)
	accessToken := signInResult.Data.Tokens.AccessToken
	assert.NotEmpty(t, accessToken)

	// 4. Get Session
	req = httptest.NewRequest("GET", "/api/v1/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var sessionResult struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			User struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sessionResult))
	assert.True(t, sessionResult.Success)
	assert.Equal(t, "Bob", sessionResult.Data.User.Name)
	assert.Equal(t, "bob@example.com", sessionResult.Data.User.Email)
}

func registerAndVerifyUser(t *testing.T, app *fiber.App, mailer *mockMailer, name, email, password string) {
	t.Helper()
	signUpReq := map[string]string{
		"name":     name,
		"email":    email,
		"password": password,
	}
	body, err := json.Marshal(signUpReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/sign-up", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)

	require.NotEmpty(t, mailer.rawCalls)
	code := extractOTPCode(t, mailer.rawCalls[len(mailer.rawCalls)-1].Body)

	verifyReq := map[string]string{
		"email": email,
		"code":  code,
	}
	body, err = json.Marshal(verifyReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/sign-up/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthEndpoints_TokenRefreshAndSignOut(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "Alice", "alice@example.com", "password123")

	// SignIn
	signInReq := map[string]string{
		"email":    "alice@example.com",
		"password": "password123",
	}
	body, err := json.Marshal(signInReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var signInResult struct {
		Data struct {
			Tokens struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signInResult))
	tokens := signInResult.Data.Tokens
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)

	// Refresh Token
	refreshReq := map[string]string{
		"refresh_token": tokens.RefreshToken,
	}
	body, err = json.Marshal(refreshReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var refreshResult struct {
		Data struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&refreshResult))
	newTokens := refreshResult.Data
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)

	// SignOut
	req = httptest.NewRequest("POST", "/api/v1/auth/sign-out", nil)
	req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Session should now fail as token is invalidated
	req = httptest.NewRequest("GET", "/api/v1/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+newTokens.AccessToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestAuthEndpoints_ChangePassword(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "Charlie", "charlie@example.com", "password123")

	// SignIn
	signInReq := map[string]string{
		"email":    "charlie@example.com",
		"password": "password123",
	}
	body, err := json.Marshal(signInReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var signInResult struct {
		Data struct {
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signInResult))
	accessToken := signInResult.Data.Tokens.AccessToken

	// Change Password
	changeReq := map[string]string{
		"current_password": "password123",
		"new_password":     "newpassword999",
	}
	body, err = json.Marshal(changeReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// SignIn with old password fails
	body, err = json.Marshal(signInReq)
	require.NoError(t, err)
	req = httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// SignIn with new password succeeds
	signInReq["password"] = "newpassword999"
	body, err = json.Marshal(signInReq)
	require.NoError(t, err)
	req = httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestAuthEndpoints_PasswordResetWorkflow(t *testing.T) {
	app, mailer, _ := newTestApp(t)
	registerAndVerifyUser(t, app, mailer, "David", "david@example.com", "password123")

	// 1. Request Reset
	resetReq := map[string]string{
		"email": "david@example.com",
	}
	body, err := json.Marshal(resetReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Extract code
	require.NotEmpty(t, mailer.rawCalls)
	code := extractOTPCode(t, mailer.rawCalls[len(mailer.rawCalls)-1].Body)

	// 2. Verify Reset
	verifyResetReq := map[string]string{
		"email": "david@example.com",
		"code":  code,
	}
	body, err = json.Marshal(verifyResetReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/reset-password/verify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var verifyResetResult struct {
		Data struct {
			ResetToken string `json:"reset_token"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&verifyResetResult))
	resetToken := verifyResetResult.Data.ResetToken
	assert.NotEmpty(t, resetToken)

	// 3. Confirm Reset
	confirmResetReq := map[string]string{
		"token":    resetToken,
		"password": "davidnewpassword123",
	}
	body, err = json.Marshal(confirmResetReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/reset-password/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	// Try to login with new password
	signInReq := map[string]string{
		"email":    "david@example.com",
		"password": "davidnewpassword123",
	}
	body, err = json.Marshal(signInReq)
	require.NoError(t, err)

	req = httptest.NewRequest("POST", "/api/v1/auth/sign-in", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func newTestApp(t *testing.T) (*fiber.App, *mockMailer, *providers.Runtime) {
	t.Helper()

	cfg := testkit.NewSQLiteConfig(t)
	cfg.Mail.Enabled = true
	cfg.Cache.Driver = "memory"
	cfg.JWT = configs.JWTConfig{
		Secret:         "test-secret-key-that-is-long-enough-to-meet-requirements-123456",
		ExpirationTime: 3600,
		RefreshTime:    7200,
		Issuer:         "lfiber",
	}
	cfg.Hash = configs.HashConfig{
		Driver: "bcrypt",
		Bcrypt: configs.BcryptHashConfig{Rounds: 4}, // Use low rounds for speed
	}

	_, conn, err := dbProvider.RegisterDatabase(cfg)
	require.NoError(t, err)

	db, err := conn.GetDB()
	require.NoError(t, err)

	createAuthTablesInSQLite(t, db)

	cacheManager := cacheProvider.NewManager(cfg)
	hasher, err := hashProvider.RegisterHash(cfg)
	require.NoError(t, err)

	_, translator, err := i18nProvider.RegisterI18n(&configs.Config{
		I18n: configs.I18nConfig{
			DefaultLanguage:    "zh-CN",
			SupportedLanguages: []string{"en", "zh-CN"},
			LanguageDir:        testkit.RepoRoot(t) + "/lang",
			CookieName:         "lang",
			CookieMaxAge:       86400,
		},
	})
	require.NoError(t, err)

	mailer := &mockMailer{}

	// Instantiate the authentic auth provider using registry
	authManager, err := authProvider.Register(cfg, conn, hasher)
	require.NoError(t, err)
	authorizer, err := authorizationProvider.Register(cfg.Authorization)
	require.NoError(t, err)

	rt := &providers.Runtime{
		Config:        cfg,
		Connection:    conn,
		Cache:         cacheManager.Store(),
		CacheManager:  cacheManager,
		Hash:          hasher,
		Translator:    translator,
		Auth:          authManager,
		Authorization: authorizer,
		EmailService:  mailer,
	}
	providers.SetInstance(rt)

	// Inject User creator callback
	rt.Auth.SetModelCreator("users", func() any {
		return &user.User{}
	})

	t.Cleanup(func() {
		providers.SetInstance(nil)
		_ = rt.Close()
	})

	app := bootstrap.NewHTTPApp(cfg)
	require.NoError(t, bootstrap.SetupApplicationRoutes(app))

	return app, mailer, rt
}

func createAuthTablesInSQLite(t *testing.T, db *sql.DB) {
	t.Helper()

	_, err := db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		avatar TEXT NULL,
		phone TEXT NULL,
		status TEXT NOT NULL DEFAULT 'active',
		email_verified_at DATETIME NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME NULL
	)`)
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE auth_otps (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL,
		purpose TEXT NOT NULL,
		code_hash TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		sent_at DATETIME NOT NULL,
		attempts INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 5,
		consumed_at DATETIME NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	require.NoError(t, err)
}

func extractOTPCode(t *testing.T, body string) string {
	t.Helper()

	re := regexp.MustCompile(`\b\d{6}\b`)
	matches := re.FindAllStringIndex(body, -1)
	for _, loc := range matches {
		start := loc[0]
		end := loc[1]
		if start > 0 && body[start-1] == '#' {
			continue
		}
		return body[start:end]
	}
	require.Fail(t, "expected otp code in email body")
	return ""
}

// mockMailer facilitates intercepting registration OTPs.
type mockMailer struct {
	rawCalls []struct {
		To      string
		Subject string
		Body    string
	}
}

func (m *mockMailer) To(to ...string) mailContracts.Message {
	return &mockMessage{mailer: m, to: to}
}

func (m *mockMailer) Send(message mailContracts.Message) error {
	msg := message.(*mockMessage)
	m.rawCalls = append(m.rawCalls, struct {
		To      string
		Subject string
		Body    string
	}{
		To:      msg.to[0],
		Subject: msg.subject,
		Body:    msg.body,
	})
	return nil
}

func (m *mockMailer) Raw(to, subject, body string) error {
	m.rawCalls = append(m.rawCalls, struct {
		To      string
		Subject string
		Body    string
	}{
		To:      to,
		Subject: subject,
		Body:    body,
	})
	return nil
}

func (m *mockMailer) Close() error       { return nil }
func (m *mockMailer) HealthCheck() error { return nil }

type mockMessage struct {
	mailer  *mockMailer
	to      []string
	subject string
	body    string
}

func (m *mockMessage) To(to ...string) mailContracts.Message        { m.to = to; return m }
func (m *mockMessage) Cc(cc ...string) mailContracts.Message        { return m }
func (m *mockMessage) Bcc(bcc ...string) mailContracts.Message      { return m }
func (m *mockMessage) Subject(subject string) mailContracts.Message { m.subject = subject; return m }

func (m *mockMessage) Html(body string) mailContracts.Message { m.body = body; return m }

func (m *mockMessage) Plain(body string) mailContracts.Message                { m.body = body; return m }
func (m *mockMessage) Attach(filePath string) mailContracts.Message           { return m }
func (m *mockMessage) Data(data map[string]interface{}) mailContracts.Message { return m }
func (m *mockMessage) View(templateName string, data map[string]interface{}) mailContracts.Message {
	return m
}

func (m *mockMessage) Mailable(ml mailContracts.Mailable) mailContracts.Message {
	m.subject = ml.Subject()
	name, data := ml.Template()
	htmlContent, err := drivers.RenderTemplate(name, data)
	if err != nil {
		if codeVal, ok := data["Code"]; ok {
			if codeStr, ok := codeVal.(string); ok {
				m.body = codeStr
			}
		}
	} else {
		m.body = htmlContent
	}
	return m
}
func (m *mockMessage) GetTo() []string                 { return m.to }
func (m *mockMessage) GetCc() []string                 { return nil }
func (m *mockMessage) GetBcc() []string                { return nil }
func (m *mockMessage) GetSubject() string              { return m.subject }
func (m *mockMessage) GetBody() string                 { return m.body }
func (m *mockMessage) IsHtml() bool                    { return false }
func (m *mockMessage) GetAttachments() []string        { return nil }
func (m *mockMessage) GetData() map[string]interface{} { return nil }
