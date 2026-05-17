package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	exceptions "fiber-starter/app/Exceptions"
	middleware "fiber-starter/app/Http/Middleware"
	models "fiber-starter/app/Models"
	cacheContracts "fiber-starter/app/Providers/Cache/Contracts"
	database "fiber-starter/app/Providers/Database/Contracts"
	hashContracts "fiber-starter/app/Providers/Hash/Contracts"
	mailContracts "fiber-starter/app/Providers/Mail/Contracts"
	repositories "fiber-starter/app/Repositories"
	helpers "fiber-starter/app/Support"
	"fiber-starter/configs"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

const (
	authOTPPurposeSignup        = models.AuthOTPPurposeSignup
	authOTPPurposePasswordReset = models.AuthOTPPurposePasswordReset

	authOTPExpiry            = 10 * time.Minute
	authOTPCooldown          = 60 * time.Second
	authOTPMaxAttempts int64 = 5

	passwordResetTokenTTL = 15 * time.Minute
)

// AuthService 认证服务接口
type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (SignUpResult, error)
	VerifySignUp(ctx context.Context, input VerifyCodeInput) (AuthResult, error)
	Login(ctx context.Context, input LoginInput) (AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (TokenPair, error)
	Logout(ctx context.Context, token string) error
	ChangePassword(ctx context.Context, input ChangePasswordInput) error
	RequestPasswordReset(ctx context.Context, input PasswordResetRequestInput) error
	VerifyPasswordReset(ctx context.Context, input VerifyCodeInput) (PasswordResetToken, error)
	ResetPassword(ctx context.Context, input ConfirmPasswordResetInput) error
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type VerifyCodeInput struct {
	Email string
	Code  string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthResult struct {
	User   *models.User
	Tokens TokenPair
}

type SignUpResult struct {
	User                 *models.User
	VerificationRequired bool
}

type ChangePasswordInput struct {
	UserID          int64
	CurrentPassword string
	NewPassword     string
}

type PasswordResetRequestInput struct {
	Email string
}

type PasswordResetToken struct {
	Token string
}

type ConfirmPasswordResetInput struct {
	Token       string
	NewPassword string
}

type authService struct {
	db     database.Connection
	config *configs.Config
	cache  cacheContracts.Store
	mailer mailContracts.Mailer
	hasher hashContracts.Hasher
}

func NewAuthService(db database.Connection, cfg *configs.Config, cache cacheContracts.Store, mailer mailContracts.Mailer, hasher hashContracts.Hasher) AuthService {
	return &authService{db: db, config: cfg, cache: cache, mailer: mailer, hasher: hasher}
}

func (s *authService) bunDB() (*bun.DB, error) {
	if s == nil || s.db == nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", errors.New("database connection not initialized"))
	}
	db, err := s.db.BunDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}
	return db, nil
}

func (s *authService) userRepo() (*repositories.UserRepository, error) {
	db, err := s.bunDB()
	if err != nil {
		return nil, err
	}
	return repositories.NewUserRepository(db), nil
}

func (s *authService) authRepo() (*repositories.AuthRepository, error) {
	db, err := s.bunDB()
	if err != nil {
		return nil, err
	}
	return repositories.NewAuthRepository(db), nil
}

func serviceUnavailable(err error) error {
	return exceptions.ServiceUnavailableWithCause("Service unavailable", err)
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (SignUpResult, error) {
	ctx = serviceContext(ctx)
	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
	}
	bunDB, err := s.bunDB()
	if err != nil {
		return SignUpResult{}, err
	}

	err = bunDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := repositories.NewUserRepository(tx)

		existing, queryErr := userRepo.GetByEmail(ctx, user.Email)
		if queryErr != nil && !errors.Is(queryErr, sql.ErrNoRows) {
			return fmt.Errorf("failed to query user: %w", queryErr)
		}

		hashedPassword, err := s.hasher.Make(user.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		now := helpers.UtcNow()
		user.Password = string(hashedPassword)
		user.Status = models.UserStatusPending
		user.EmailVerifiedAt = nil
		user.UpdatedAt = now

		if errors.Is(queryErr, sql.ErrNoRows) {
			user.CreatedAt = now
			if err := userRepo.Create(ctx, user); err != nil {
				return fmt.Errorf("failed to create user: %w", err)
			}
			return nil
		}

		if existing.Status == models.UserStatusActive || existing.EmailVerifiedAt != nil {
			return errors.New("email already registered")
		}

		if existing.Status == models.UserStatusSuspended || existing.Status == models.UserStatusBanned || existing.Status == models.UserStatusInactive {
			return errors.New("email already registered")
		}

		existing.Name = user.Name
		existing.Email = user.Email
		existing.Password = user.Password
		existing.Status = user.Status
		existing.EmailVerifiedAt = nil
		existing.UpdatedAt = user.UpdatedAt
		if err := userRepo.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})
	if err != nil {
		return SignUpResult{}, err
	}

	saved, err := s.getUserByEmail(ctx, user.Email)
	if err != nil {
		return SignUpResult{}, err
	}

	code, record, shouldSend, err := s.issueOTP(saved.Email, authOTPPurposeSignup)
	if err != nil {
		return SignUpResult{}, err
	}
	if shouldSend {
		if err := s.sendOTPEmail(saved.Email, "Verify your email", "Your verification code is", code, authOTPExpiry); err != nil {
			_ = s.deleteOTP(record.ID)
			return SignUpResult{}, err
		}
	}

	return SignUpResult{User: saved, VerificationRequired: true}, nil
}

func (s *authService) VerifySignUp(ctx context.Context, input VerifyCodeInput) (AuthResult, error) {
	ctx = serviceContext(ctx)
	bunDB, err := s.bunDB()
	if err != nil {
		return AuthResult{}, err
	}

	err = bunDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := repositories.NewUserRepository(tx)
		authRepo := repositories.NewAuthRepository(tx)

		otp, err := s.validateOTP(authRepo, input.Email, authOTPPurposeSignup, input.Code)
		if err != nil {
			return err
		}

		user, err := userRepo.GetByEmail(ctx, input.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
			return fmt.Errorf("failed to query user: %w", err)
		}

		if user.Status == models.UserStatusBanned || user.Status == models.UserStatusSuspended || user.Status == models.UserStatusInactive {
			return errors.New("user account has been disabled")
		}

		now := helpers.UtcNow()
		user.Status = models.UserStatusActive
		user.EmailVerifiedAt = &now
		user.UpdatedAt = now
		if err := userRepo.Update(ctx, user); err != nil {
			return fmt.Errorf("failed to activate user: %w", err)
		}

		if err := authRepo.ConsumeOTP(ctx, otp.ID, now, now); err != nil {
			return fmt.Errorf("failed to consume otp: %w", err)
		}
		return nil
	})
	if err != nil {
		return AuthResult{}, err
	}

	user, err := s.getUserByEmail(ctx, input.Email)
	if err != nil {
		return AuthResult{}, err
	}

	accessToken, refreshToken, err := s.issueSessionTokens(user)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: user,
		Tokens: TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	ctx = serviceContext(ctx)
	userRepo, err := s.userRepo()
	if err != nil {
		return AuthResult{}, err
	}

	user, err := userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuthResult{}, errors.New("invalid email or password")
		}
		helpers.LogError("Failed to query user", zap.Error(err))
		return AuthResult{}, fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status == models.UserStatusPending {
		return AuthResult{}, errors.New("email verification required")
	}
	if user.Status == models.UserStatusInactive || user.Status == models.UserStatusSuspended || user.Status == models.UserStatusBanned {
		return AuthResult{}, errors.New("user account has been disabled")
	}

	if ok := s.hasher.Check(input.Password, user.Password); !ok {
		return AuthResult{}, errors.New("invalid email or password")
	}

	accessToken, refreshToken, err := s.issueSessionTokens(user)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User: user,
		Tokens: TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (TokenPair, error) {
	ctx = serviceContext(ctx)
	claims, err := middleware.ValidateToken(refreshToken, s.config)
	if err != nil {
		return TokenPair{}, errors.New("invalid refresh token")
	}

	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	cachedToken, err := s.cache.Get(cacheKey)
	if err != nil {
		return TokenPair{}, serviceUnavailable(err)
	}
	if cachedToken != refreshToken {
		return TokenPair{}, errors.New("refresh token has expired")
	}

	user, err := s.getUserByID(ctx, claims.UserID)
	if err != nil {
		return TokenPair{}, err
	}
	if user.Status == models.UserStatusPending {
		return TokenPair{}, errors.New("email verification required")
	}
	if user.Status == models.UserStatusInactive || user.Status == models.UserStatusSuspended || user.Status == models.UserStatusBanned {
		return TokenPair{}, errors.New("user account has been disabled")
	}

	accessToken, newRefreshToken, err := s.issueSessionTokens(user)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	_ = serviceContext(ctx)
	claims, err := middleware.ValidateToken(token, s.config)
	if err != nil {
		return errors.New("invalid token")
	}

	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	if err := s.cache.Delete(cacheKey); err != nil {
		return serviceUnavailable(err)
	}

	blacklistKey := fmt.Sprintf("blacklist:%s", token)
	if err := s.cache.Set(blacklistKey, "1", time.Duration(s.config.JWT.ExpirationTime)*time.Second); err != nil {
		return serviceUnavailable(err)
	}

	return nil
}

func (s *authService) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	ctx = serviceContext(ctx)
	userRepo, err := s.userRepo()
	if err != nil {
		return err
	}

	user, err := userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return fmt.Errorf("user not found: %w", err)
	}

	if ok := s.hasher.Check(input.CurrentPassword, user.Password); !ok {
		return errors.New("incorrect current password")
	}

	hashedPassword, err := s.hasher.Make(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := userRepo.UpdatePassword(ctx, input.UserID, string(hashedPassword), helpers.UtcNow()); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *authService) RequestPasswordReset(ctx context.Context, input PasswordResetRequestInput) error {
	ctx = serviceContext(ctx)
	userRepo, err := s.userRepo()
	if err != nil {
		return err
	}

	user, err := userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	if user.Status == models.UserStatusInactive || user.Status == models.UserStatusSuspended || user.Status == models.UserStatusBanned {
		return nil
	}

	code, record, shouldSend, err := s.issueOTP(input.Email, authOTPPurposePasswordReset)
	if err != nil {
		return err
	}
	if shouldSend {
		if err := s.sendOTPEmail(input.Email, "Reset your password", "Your password reset code is", code, authOTPExpiry); err != nil {
			_ = s.deleteOTP(record.ID)
			return err
		}
	}
	return nil
}

func (s *authService) VerifyPasswordReset(ctx context.Context, input VerifyCodeInput) (PasswordResetToken, error) {
	var resetToken string
	ctx = serviceContext(ctx)
	bunDB, err := s.bunDB()
	if err != nil {
		return PasswordResetToken{}, err
	}

	err = bunDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		userRepo := repositories.NewUserRepository(tx)
		authRepo := repositories.NewAuthRepository(tx)

		otp, err := s.validateOTP(authRepo, input.Email, authOTPPurposePasswordReset, input.Code)
		if err != nil {
			return err
		}

		user, err := userRepo.GetByEmail(ctx, input.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return fmt.Errorf("failed to query user: %w", err)
		}

		if user.Status == models.UserStatusInactive || user.Status == models.UserStatusSuspended || user.Status == models.UserStatusBanned {
			return nil
		}

		now := helpers.UtcNow()
		if err := authRepo.ConsumeOTP(ctx, otp.ID, now, now); err != nil {
			return fmt.Errorf("failed to consume otp: %w", err)
		}

		token, err := s.makePasswordResetToken(user, now.Add(passwordResetTokenTTL))
		if err != nil {
			return err
		}
		cacheKey := fmt.Sprintf("password_reset_token:%s", token)
		if err := s.cache.Set(cacheKey, fmt.Sprintf("%d|%s", user.ID, user.Email), passwordResetTokenTTL); err != nil {
			return serviceUnavailable(err)
		}
		resetToken = token
		return nil
	})
	if err != nil {
		return PasswordResetToken{}, err
	}

	if resetToken == "" {
		return PasswordResetToken{}, errors.New("failed to create password reset token")
	}

	return PasswordResetToken{Token: resetToken}, nil
}

func (s *authService) ResetPassword(ctx context.Context, input ConfirmPasswordResetInput) error {
	userID, email, err := s.parsePasswordResetToken(input.Token)
	if err != nil {
		return err
	}

	ctx = serviceContext(ctx)
	userRepo, err := s.userRepo()
	if err != nil {
		return err
	}

	user, err := userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		return fmt.Errorf("user not found: %w", err)
	}
	if user.Email != email {
		return errors.New("invalid reset token")
	}

	cacheKey := fmt.Sprintf("password_reset_token:%s", input.Token)
	cachedValue, err := s.cache.Get(cacheKey)
	if err != nil {
		return serviceUnavailable(err)
	}
	if cachedValue != fmt.Sprintf("%d|%s", user.ID, user.Email) {
		return errors.New("invalid reset token")
	}

	hashedPassword, err := s.hasher.Make(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword), helpers.UtcNow()); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	refreshCacheKey := fmt.Sprintf("refresh_token:%d", user.ID)
	if err := s.cache.Delete(refreshCacheKey); err != nil {
		return serviceUnavailable(err)
	}
	if err := s.cache.Delete(fmt.Sprintf("password_reset_token:%s", input.Token)); err != nil {
		return serviceUnavailable(err)
	}

	return nil
}

func (s *authService) issueSessionTokens(user *models.User) (string, string, error) {
	accessToken, err := middleware.GenerateToken(user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	cacheKey := fmt.Sprintf("refresh_token:%d", user.ID)
	if err := s.cache.Set(cacheKey, refreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		return "", "", serviceUnavailable(err)
	}

	return accessToken, refreshToken, nil
}

func (s *authService) issueOTP(email string, purpose models.AuthOTPPurpose) (string, *models.AuthOTP, bool, error) {
	now := helpers.UtcNow()

	latest, err := s.getLatestOTP(email, purpose)
	if err == nil && now.Sub(latest.SentAt) < authOTPCooldown && latest.ExpiresAt.After(now) {
		return "", latest, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", nil, false, err
	}

	code, err := generateOTPCode()
	if err != nil {
		return "", nil, false, err
	}

	record, err := s.createOTP(email, purpose, code, now)
	if err != nil {
		return "", nil, false, err
	}

	return code, record, true, nil
}

func (s *authService) createOTP(email string, purpose models.AuthOTPPurpose, code string, now time.Time) (*models.AuthOTP, error) {
	ctx := context.Background()
	authRepo, err := s.authRepo()
	if err != nil {
		return nil, err
	}

	otp := &models.AuthOTP{
		Email:       email,
		Purpose:     purpose,
		CodeHash:    hashOTPCode(s.otpSecret(), email, purpose, code),
		ExpiresAt:   now.Add(authOTPExpiry),
		SentAt:      now,
		Attempts:    0,
		MaxAttempts: authOTPMaxAttempts,
		ConsumedAt:  nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := authRepo.CreateOTP(ctx, otp); err != nil {
		return nil, err
	}
	return otp, nil
}

func (s *authService) deleteOTP(id int64) error {
	ctx := context.Background()
	authRepo, err := s.authRepo()
	if err != nil {
		return err
	}
	return authRepo.DeleteOTPByID(ctx, id)
}

func (s *authService) getLatestOTP(email string, purpose models.AuthOTPPurpose) (*models.AuthOTP, error) {
	ctx := context.Background()
	authRepo, err := s.authRepo()
	if err != nil {
		return nil, err
	}
	return authRepo.GetLatestOTPByEmailPurpose(ctx, email, purpose)
}

func (s *authService) validateOTP(authRepo *repositories.AuthRepository, email string, purpose models.AuthOTPPurpose, code string) (*models.AuthOTP, error) {
	ctx := context.Background()
	otp, err := authRepo.GetLatestOTPByEmailPurpose(ctx, email, purpose)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid verification code")
		}
		return nil, fmt.Errorf("failed to query otp: %w", err)
	}

	now := helpers.UtcNow()
	if now.After(otp.ExpiresAt) {
		if err := authRepo.ConsumeOTP(ctx, otp.ID, now, now); err != nil {
			return nil, fmt.Errorf("failed to consume expired otp: %w", err)
		}
		return nil, errors.New("verification code has expired")
	}

	if !equalOTP(hashOTPCode(s.otpSecret(), email, purpose, code), otp.CodeHash) {
		if err := authRepo.IncrementOTPAttempts(ctx, otp.ID, now); err != nil {
			return nil, fmt.Errorf("failed to update otp attempts: %w", err)
		}
		if otp.Attempts+1 >= otp.MaxAttempts {
			if err := authRepo.ConsumeOTP(ctx, otp.ID, now, now); err != nil {
				return nil, fmt.Errorf("failed to lock otp: %w", err)
			}
			return nil, errors.New("verification attempts exceeded")
		}
		return nil, errors.New("invalid verification code")
	}

	return otp, nil
}

func (s *authService) getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx = serviceContext(ctx)
	userRepo, err := s.userRepo()
	if err != nil {
		return nil, err
	}
	return userRepo.GetByEmail(ctx, email)
}

func (s *authService) getUserByID(ctx context.Context, id int64) (*models.User, error) {
	ctx = serviceContext(ctx)
	userRepo, err := s.userRepo()
	if err != nil {
		return nil, err
	}
	return userRepo.GetByID(ctx, id)
}

func (s *authService) makePasswordResetToken(user *models.User, expiresAt time.Time) (string, error) {
	if user == nil {
		return "", errors.New("user is required")
	}

	payload := fmt.Sprintf("%d|%s|%d", user.ID, user.Email, expiresAt.Unix())
	mac := hmac.New(sha256.New, []byte(s.otpSecret()))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", err
	}

	raw := payload + "|" + hex.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(raw)), nil
}

func (s *authService) parsePasswordResetToken(token string) (int64, string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, "", errors.New("invalid reset token")
	}

	parts := strings.Split(string(decoded), "|")
	if len(parts) != 4 {
		return 0, "", errors.New("invalid reset token")
	}

	expiresUnix, err := parseInt64(parts[2])
	if err != nil {
		return 0, "", errors.New("invalid reset token")
	}
	if helpers.UtcNow().Unix() > expiresUnix {
		return 0, "", errors.New("reset token has expired")
	}

	mac := hmac.New(sha256.New, []byte(s.otpSecret()))
	payload := strings.Join(parts[:3], "|")
	if _, err := mac.Write([]byte(payload)); err != nil {
		return 0, "", err
	}

	expectedSig := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(parts[3])) != 1 {
		return 0, "", errors.New("invalid reset token")
	}

	userID, err := parseInt64(parts[0])
	if err != nil {
		return 0, "", errors.New("invalid reset token")
	}

	return userID, parts[1], nil
}

func (s *authService) otpSecret() string {
	if s != nil && s.config != nil && strings.TrimSpace(s.config.JWT.Secret) != "" {
		return s.config.JWT.Secret
	}
	if s != nil && s.config != nil && strings.TrimSpace(s.config.App.Name) != "" {
		return s.config.App.Name
	}
	return "fiber-starter-otp-secret"
}

func (s *authService) sendOTPEmail(to, subject, leadText, code string, expires time.Duration) error {
	if s.mailer == nil {
		return errors.New("mailer is not configured")
	}

	body := fmt.Sprintf(
		`<div style="font-family:Arial,sans-serif;line-height:1.6">
<p>%s</p>
<p><strong style="font-size:24px;letter-spacing:4px">%s</strong></p>
<p>This code expires in %d minutes.</p>
</div>`,
		leadText,
		code,
		int(expires.Minutes()),
	)

	return s.mailer.Raw(to, subject, body)
}

func generateOTPCode() (string, error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func hashOTPCode(secret, email string, purpose models.AuthOTPPurpose, code string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{secret, email, string(purpose), code}, "|")))
	return hex.EncodeToString(sum[:])
}

func equalOTP(left, right string) bool {
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}

func parseInt64(value string) (int64, error) {
	return strconv.ParseInt(value, 10, 64)
}
