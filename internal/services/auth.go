// Package services 提供应用程序的业务逻辑服务
package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"fiber-starter/internal/config"
	database "fiber-starter/internal/db"
	models "fiber-starter/internal/domain/model"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/transport/http/middleware"

	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/scan"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(user *models.User) error
	Login(email, password string) (*models.User, string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(token string) error
	ChangePassword(userID int64, oldPassword, newPassword string) error
	ForgotPassword(email string) error
	ResetPassword(token, email, newPassword string) error
}

// authService 认证服务实现
type authService struct {
	db     *database.Connection
	config *config.Config
	cache  helpers.CacheService
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *database.Connection, cfg *config.Config, cache helpers.CacheService) AuthService {
	return &authService{
		db:     db,
		config: cfg,
		cache:  cache,
	}
}

// Register User registration
func (s *authService) Register(user *models.User) error {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return err
	}

	exists, err := userExistsByEmail(ctx, db, dialect, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	q, err := insertUserQuery(dialect, *user)
	if err != nil {
		return err
	}
	if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Login User login
func (s *authService) Login(email, password string) (*models.User, string, string, error) {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return nil, "", "", err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return nil, "", "", err
	}

	u, err := getUserByEmail(ctx, db, dialect, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", errors.New("invalid email or password")
		}
		helpers.LogError("Failed to query user", zap.Error(err))
		return nil, "", "", fmt.Errorf("failed to query user: %w", err)
	}

	// Check user status
	if !u.IsActive() {
		return nil, "", "", errors.New("user account has been disabled")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	// Generate JWT tokens
	accessToken, err := middleware.GenerateToken(&u, s.config)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(&u, s.config)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in cache
	cacheKey := fmt.Sprintf("refresh_token:%d", u.ID)
	if err := s.cache.Set(cacheKey, refreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		// Cache failure doesn't affect login, just log it
		helpers.LogError("Failed to cache refresh token", zap.Error(err))
	}

	return &u, accessToken, refreshToken, nil
}

// RefreshToken Refresh access token
func (s *authService) RefreshToken(refreshToken string) (string, string, error) {
	// Validate refresh token
	claims, err := middleware.ValidateToken(refreshToken, s.config)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Check refresh token in cache
	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	cachedToken, err := s.cache.Get(cacheKey)
	if err != nil || cachedToken != refreshToken {
		return "", "", errors.New("refresh token has expired")
	}

	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return "", "", err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return "", "", err
	}

	user, err := getUserByID(ctx, db, dialect, claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	// Check user status
	if !user.IsActive() {
		return "", "", errors.New("user account has been disabled")
	}

	// Generate new tokens
	newAccessToken, err := middleware.GenerateToken(&user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := middleware.GenerateRefreshToken(&user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Update refresh token in cache
	if err := s.cache.Set(cacheKey, newRefreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		helpers.LogError("Failed to update refresh token cache", zap.Error(err))
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout User logout
func (s *authService) Logout(token string) error {
	// Validate token to get user ID
	claims, err := middleware.ValidateToken(token, s.config)
	if err != nil {
		return errors.New("invalid token")
	}

	// Delete refresh token from cache
	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	if err := s.cache.Delete(cacheKey); err != nil {
		helpers.LogError("Failed to delete refresh token cache", zap.Error(err))
	}

	// Add access token to blacklist (cache)
	blacklistKey := fmt.Sprintf("blacklist:%s", token)
	if err := s.cache.Set(blacklistKey, "1", time.Duration(s.config.JWT.ExpirationTime)*time.Second); err != nil {
		helpers.LogError("Failed to blacklist token", zap.Error(err))
	}

	return nil
}

// ChangePassword Change password
func (s *authService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return err
	}

	user, err := getUserByID(ctx, db, dialect, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	q, err := updateUserPasswordQuery(dialect, userID, string(hashedPassword), time.Now().UTC())
	if err != nil {
		return err
	}
	if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ForgotPassword Forgot password
func (s *authService) ForgotPassword(email string) error {
	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return err
	}

	user, err := getUserByEmail(ctx, db, dialect, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	// Generate reset token
	resetToken := fmt.Sprintf("%d:%d", user.ID, time.Now().Unix())

	// Store reset token in cache (valid for 1 hour)
	cacheKey := fmt.Sprintf("reset_token:%s", resetToken)
	if err := s.cache.Set(cacheKey, user.Email, time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// TODO: Send reset password email
	helpers.Info("Reset password token", zap.String("token", resetToken), zap.String("email", user.Email))

	return nil
}

// ResetPassword Reset password
func (s *authService) ResetPassword(token, email, newPassword string) error {
	// Validate reset token
	cacheKey := fmt.Sprintf("reset_token:%s", token)
	cachedEmail, err := s.cache.Get(cacheKey)
	if err != nil || cachedEmail != email {
		return errors.New("invalid or expired reset token")
	}

	ctx := context.Background()
	db, err := s.db.GetDB()
	if err != nil {
		return err
	}

	dialect, err := s.db.Dialect()
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	q, err := updateUserPasswordByEmailQuery(dialect, email, string(hashedPassword), time.Now().UTC())
	if err != nil {
		return err
	}
	if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	// Delete reset token
	_ = s.cache.Delete(cacheKey)

	return nil
}

func rawQuery(dialect string, q string, args ...any) (bob.Query, error) {
	switch dialect {
	case "psql":
		return psql.RawQuery(q, args...), nil
	case "sqlite":
		return sqlite.RawQuery(q, args...), nil
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func userExistsByEmail(ctx context.Context, db *sql.DB, dialect string, email string) (bool, error) {
	q, err := rawQuery(dialect, "SELECT 1 FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1", email)
	if err != nil {
		return false, err
	}

	_, err = bob.One(ctx, bob.NewDB(db), q, scan.SingleColumnMapper[int])
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func getUserByEmail(ctx context.Context, db *sql.DB, dialect string, email string) (models.User, error) {
	q, err := rawQuery(dialect, "SELECT id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at FROM users WHERE email = ? AND deleted_at IS NULL LIMIT 1", email)
	if err != nil {
		return models.User{}, err
	}
	return bob.One(ctx, bob.NewDB(db), q, scan.StructMapper[models.User]())
}

func getUserByID(ctx context.Context, db *sql.DB, dialect string, id int64) (models.User, error) {
	q, err := rawQuery(dialect, "SELECT id, name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at FROM users WHERE id = ? AND deleted_at IS NULL LIMIT 1", id)
	if err != nil {
		return models.User{}, err
	}
	return bob.One(ctx, bob.NewDB(db), q, scan.StructMapper[models.User]())
}

func insertUserQuery(dialect string, user models.User) (bob.Query, error) {
	q := "INSERT INTO users (name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	return rawQuery(dialect, q,
		user.Name,
		user.Email,
		user.Password,
		user.Avatar,
		user.Phone,
		string(user.Status),
		user.EmailVerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
		user.DeletedAt,
	)
}

func updateUserPasswordQuery(dialect string, userID int64, hashedPassword string, now time.Time) (bob.Query, error) {
	q := "UPDATE users SET password = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL"
	return rawQuery(dialect, q, hashedPassword, now, userID)
}

func updateUserPasswordByEmailQuery(dialect string, email string, hashedPassword string, now time.Time) (bob.Query, error) {
	q := "UPDATE users SET password = ?, updated_at = ? WHERE email = ? AND deleted_at IS NULL"
	return rawQuery(dialect, q, hashedPassword, now, email)
}
