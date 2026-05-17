package services_test

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"testing"
	"time"

	"fiber-starter/configs"
	auth "fiber-starter/internal/features/auth"
	user "fiber-starter/internal/features/user"
	cacheContracts "fiber-starter/internal/providers/cache/contracts"
	dbProvider "fiber-starter/internal/providers/database"
	mailContracts "fiber-starter/internal/providers/mail/contracts"
	"fiber-starter/tests/internal/testkit"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_SignUpVerifyAndLoginWorkflow(t *testing.T) {
	svc, mailer, cache := newAuthServiceTestHarness(t)
	ctx := context.Background()

	signUp, err := svc.Register(ctx, auth.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)
	dbUser := signUp.User
	require.NotNil(t, dbUser)
	assert.Equal(t, user.UserStatusPending, dbUser.Status)
	assert.Nil(t, dbUser.EmailVerifiedAt)
	require.Len(t, mailer.rawCalls, 1)

	_, err = svc.Login(ctx, auth.LoginInput{Email: "alice@example.com", Password: "password123"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email verification required")

	code := extractCode(t, mailer.rawCalls[0].Body)
	authResult, err := svc.VerifySignUp(ctx, auth.VerifyCodeInput{Email: "alice@example.com", Code: code})
	require.NoError(t, err)
	verifiedUser := authResult.User
	require.NotNil(t, verifiedUser)
	assert.Equal(t, user.UserStatusActive, verifiedUser.Status)
	require.NotNil(t, verifiedUser.EmailVerifiedAt)
	assert.NotEmpty(t, authResult.Tokens.AccessToken)
	assert.NotEmpty(t, authResult.Tokens.RefreshToken)

	loginResult, err := svc.Login(ctx, auth.LoginInput{Email: "alice@example.com", Password: "password123"})
	require.NoError(t, err)
	loginUser := loginResult.User
	assert.Equal(t, verifiedUser.Email, loginUser.Email)
	assert.NotEmpty(t, loginResult.Tokens.AccessToken)
	assert.NotEmpty(t, loginResult.Tokens.RefreshToken)

	storedRefresh, err := cache.Get("refresh_token:" + fmt.Sprint(loginUser.ID))
	require.NoError(t, err)
	assert.Equal(t, loginResult.Tokens.RefreshToken, storedRefresh)
}

func TestAuthService_PasswordResetOTPWorkflow(t *testing.T) {
	svc, mailer, cache := newAuthServiceTestHarness(t)
	ctx := context.Background()

	seedVerifiedUser(t, svc, mailer)
	mailer.rawCalls = nil

	require.NoError(t, svc.RequestPasswordReset(ctx, auth.PasswordResetRequestInput{Email: "alice@example.com"}))
	require.Len(t, mailer.rawCalls, 1)

	code := extractCode(t, mailer.rawCalls[0].Body)
	resetToken, err := svc.VerifyPasswordReset(ctx, auth.VerifyCodeInput{Email: "alice@example.com", Code: code})
	require.NoError(t, err)
	require.NotEmpty(t, resetToken.Token)

	require.NoError(t, svc.ResetPassword(ctx, auth.ConfirmPasswordResetInput{Token: resetToken.Token, NewPassword: "new-password-123"}))
	require.Error(t, svc.ResetPassword(ctx, auth.ConfirmPasswordResetInput{Token: resetToken.Token, NewPassword: "new-password-456"}))

	_, err = svc.Login(ctx, auth.LoginInput{Email: "alice@example.com", Password: "password123"})
	require.Error(t, err)

	loginResult, err := svc.Login(ctx, auth.LoginInput{Email: "alice@example.com", Password: "new-password-123"})
	require.NoError(t, err)
	dbUser := loginResult.User
	assert.Equal(t, "alice@example.com", dbUser.Email)
	assert.NotEmpty(t, loginResult.Tokens.AccessToken)
	assert.NotEmpty(t, loginResult.Tokens.RefreshToken)

	exists, err := cache.Exists("refresh_token:" + fmt.Sprint(dbUser.ID))
	require.NoError(t, err)
	assert.True(t, exists)
}

func newAuthServiceTestHarness(t *testing.T) (auth.AuthService, *fakeMailer, *fakeCache) {
	t.Helper()

	cfg := &configs.Config{
		App: configs.AppConfig{
			Name: "fiber-starter",
		},
		JWT: configs.JWTConfig{
			Secret:         "test-secret",
			ExpirationTime: 3600,
			RefreshTime:    7200,
			Issuer:         "fiber-starter",
		},
		Hash: configs.HashConfig{
			Driver: "bcrypt",
			Bcrypt: configs.BcryptHashConfig{Rounds: 4}, // Use low rounds for speed in tests
		},
	}

	_, conn, err := dbProvider.RegisterDatabase(testkit.NewSQLiteConfig(t))
	require.NoError(t, err)

	db, err := conn.GetDB()
	require.NoError(t, err)

	createAuthTables(t, db)

	cache := newFakeCache()
	mailer := &fakeMailer{}
	svc := auth.NewAuthService(conn, cfg, cache, mailer)

	return svc, mailer, cache
}

func createAuthTables(t *testing.T, db *sql.DB) {
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
		deleted_at DATETIME NULL,
		CONSTRAINT ck_users_status CHECK (status IN ('active','inactive','pending','suspended','banned'))
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
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CONSTRAINT ck_auth_otps_purpose CHECK (purpose IN ('signup','password_reset'))
	)`)
	require.NoError(t, err)
}

func seedVerifiedUser(t *testing.T, svc auth.AuthService, mailer *fakeMailer) {
	t.Helper()
	ctx := context.Background()

	_, err := svc.Register(ctx, auth.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	require.Len(t, mailer.rawCalls, 1)
	code := extractCode(t, mailer.rawCalls[0].Body)
	_, err = svc.VerifySignUp(ctx, auth.VerifyCodeInput{Email: "alice@example.com", Code: code})
	require.NoError(t, err)
}

func extractCode(t *testing.T, body string) string {
	t.Helper()

	re := regexp.MustCompile(`\b\d{6}\b`)
	match := re.FindString(body)
	require.NotEmpty(t, match, "expected otp code in email body")
	return match
}

type fakeMailer struct {
	rawCalls []fakeRawCall
}

type fakeRawCall struct {
	To      string
	Subject string
	Body    string
}

func (m *fakeMailer) To(to ...string) mailContracts.Message    { return fakeMessage{} }
func (m *fakeMailer) Send(message mailContracts.Message) error { return nil }
func (m *fakeMailer) Raw(to, subject, body string) error {
	m.rawCalls = append(m.rawCalls, fakeRawCall{To: to, Subject: subject, Body: body})
	return nil
}
func (m *fakeMailer) Close() error       { return nil }
func (m *fakeMailer) HealthCheck() error { return nil }

type fakeMessage struct{}

func (fakeMessage) To(to ...string) mailContracts.Message                  { return fakeMessage{} }
func (fakeMessage) Cc(cc ...string) mailContracts.Message                  { return fakeMessage{} }
func (fakeMessage) Bcc(bcc ...string) mailContracts.Message                { return fakeMessage{} }
func (fakeMessage) Subject(subject string) mailContracts.Message           { return fakeMessage{} }
func (fakeMessage) Html(body string) mailContracts.Message                 { return fakeMessage{} }
func (fakeMessage) Plain(body string) mailContracts.Message                { return fakeMessage{} }
func (fakeMessage) Attach(filePath string) mailContracts.Message           { return fakeMessage{} }
func (fakeMessage) Data(data map[string]interface{}) mailContracts.Message { return fakeMessage{} }
func (fakeMessage) GetTo() []string                                        { return nil }
func (fakeMessage) GetCc() []string                                        { return nil }
func (fakeMessage) GetBcc() []string                                       { return nil }
func (fakeMessage) GetSubject() string                                     { return "" }
func (fakeMessage) GetBody() string                                        { return "" }
func (fakeMessage) IsHtml() bool                                           { return false }
func (fakeMessage) GetAttachments() []string                               { return nil }
func (fakeMessage) GetData() map[string]interface{}                        { return nil }

type fakeCache struct {
	values map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{values: make(map[string]string)}
}

func (c *fakeCache) Set(key string, value interface{}, expiration time.Duration) error {
	c.values[key] = fmt.Sprint(value)
	return nil
}

func (c *fakeCache) Put(key string, value interface{}, expiration time.Duration) error {
	return c.Set(key, value, expiration)
}

func (c *fakeCache) Get(key string) (string, error) {
	value, ok := c.values[key]
	if !ok {
		return "", fmt.Errorf("cache key %s not found", key)
	}
	return value, nil
}

func (c *fakeCache) GetBytes(key string) ([]byte, error) {
	value, err := c.Get(key)
	return []byte(value), err
}

func (c *fakeCache) GetJSON(key string, dest interface{}) error { return nil }
func (c *fakeCache) Delete(key string) error {
	delete(c.values, key)
	return nil
}
func (c *fakeCache) DeletePattern(pattern string) error { return nil }
func (c *fakeCache) Exists(key string) (bool, error) {
	_, ok := c.values[key]
	return ok, nil
}
func (c *fakeCache) Has(key string) (bool, error)                      { return c.Exists(key) }
func (c *fakeCache) TTL(key string) (time.Duration, error)             { return time.Minute, nil }
func (c *fakeCache) Expire(key string, expiration time.Duration) error { return nil }
func (c *fakeCache) Add(key string, value interface{}, expiration time.Duration) (bool, error) {
	if _, ok := c.values[key]; ok {
		return false, nil
	}
	c.values[key] = fmt.Sprint(value)
	return true, nil
}

func (c *fakeCache) Forever(key string, value interface{}) error {
	c.values[key] = fmt.Sprint(value)
	return nil
}
func (c *fakeCache) Forget(key string) error { return c.Delete(key) }
func (c *fakeCache) Flush() error {
	c.values = make(map[string]string)
	return nil
}

func (c *fakeCache) Pull(key string) (string, error) {
	value, err := c.Get(key)
	if err != nil {
		return "", err
	}
	_ = c.Delete(key)
	return value, nil
}
func (c *fakeCache) Increment(key string) (int64, error) { return 0, nil }
func (c *fakeCache) Decrement(key string) (int64, error) { return 0, nil }
func (c *fakeCache) Close() error                        { return nil }
func (c *fakeCache) HealthCheck() error                  { return nil }

var _ cacheContracts.Store = (*fakeCache)(nil)
