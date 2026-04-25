// Package services 提供应用程序的业务逻辑服务
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"fiber-starter/internal/config"
	database "fiber-starter/internal/db"
	models "fiber-starter/internal/domain/model"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/transport/http/middleware"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	db, err := s.db.GetGormDB()
	if err != nil {
		return err
	}

	exists, err := userExistsByEmail(context.Background(), db, user.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user.Password = string(hashedPassword)

	now := time.Now().UTC()
	if user.Status == "" {
		user.Status = models.UserStatusActive
	}
	user.CreatedAt = now
	user.UpdatedAt = now

	if err := db.WithContext(context.Background()).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// Login User login
func (s *authService) Login(email, password string) (*models.User, string, string, error) {
	db, err := s.db.GetGormDB()
	if err != nil {
		return nil, "", "", err
	}

	u, err := getUserByEmail(context.Background(), db, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errors.New("invalid email or password")
		}
		helpers.LogError("Failed to query user", zap.Error(err))
		return nil, "", "", fmt.Errorf("failed to query user: %w", err)
	}

	if !u.IsActive() {
		return nil, "", "", errors.New("user account has been disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", "", errors.New("invalid email or password")
	}

	accessToken, err := middleware.GenerateToken(&u, s.config)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(&u, s.config)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	cacheKey := fmt.Sprintf("refresh_token:%d", u.ID)
	if err := s.cache.Set(cacheKey, refreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		helpers.LogError("Failed to cache refresh token", zap.Error(err))
	}

	return &u, accessToken, refreshToken, nil
}

// RefreshToken Refresh access token
func (s *authService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := middleware.ValidateToken(refreshToken, s.config)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	cachedToken, err := s.cache.Get(cacheKey)
	if err != nil || cachedToken != refreshToken {
		return "", "", errors.New("refresh token has expired")
	}

	db, err := s.db.GetGormDB()
	if err != nil {
		return "", "", err
	}

	user, err := getUserByID(context.Background(), db, claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	if !user.IsActive() {
		return "", "", errors.New("user account has been disabled")
	}

	newAccessToken, err := middleware.GenerateToken(&user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := middleware.GenerateRefreshToken(&user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	if err := s.cache.Set(cacheKey, newRefreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		helpers.LogError("Failed to update refresh token cache", zap.Error(err))
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout User logout
func (s *authService) Logout(token string) error {
	claims, err := middleware.ValidateToken(token, s.config)
	if err != nil {
		return errors.New("invalid token")
	}

	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	if err := s.cache.Delete(cacheKey); err != nil {
		helpers.LogError("Failed to delete refresh token cache", zap.Error(err))
	}

	blacklistKey := fmt.Sprintf("blacklist:%s", token)
	if err := s.cache.Set(blacklistKey, "1", time.Duration(s.config.JWT.ExpirationTime)*time.Second); err != nil {
		helpers.LogError("Failed to blacklist token", zap.Error(err))
	}

	return nil
}

// ChangePassword Change password
func (s *authService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	db, err := s.db.GetGormDB()
	if err != nil {
		return err
	}

	user, err := getUserByID(context.Background(), db, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("incorrect current password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = db.WithContext(context.Background()).
		Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]interface{}{
			"password":   string(hashedPassword),
			"updated_at": time.Now().UTC(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ForgotPassword Forgot password
func (s *authService) ForgotPassword(email string) error {
	db, err := s.db.GetGormDB()
	if err != nil {
		return err
	}

	user, err := getUserByEmail(context.Background(), db, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	resetToken := fmt.Sprintf("%d:%d", user.ID, time.Now().Unix())

	cacheKey := fmt.Sprintf("reset_token:%s", resetToken)
	if err := s.cache.Set(cacheKey, user.Email, time.Hour); err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	helpers.Info("Reset password token", zap.String("token", resetToken), zap.String("email", user.Email))
	return nil
}

// ResetPassword Reset password
func (s *authService) ResetPassword(token, email, newPassword string) error {
	cacheKey := fmt.Sprintf("reset_token:%s", token)
	cachedEmail, err := s.cache.Get(cacheKey)
	if err != nil || cachedEmail != email {
		return errors.New("invalid or expired reset token")
	}

	db, err := s.db.GetGormDB()
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = db.WithContext(context.Background()).
		Model(&models.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Updates(map[string]interface{}{
			"password":   string(hashedPassword),
			"updated_at": time.Now().UTC(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	_ = s.cache.Delete(cacheKey)
	return nil
}

func userExistsByEmail(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error
	return count > 0, err
}

func getUserByEmail(ctx context.Context, db *gorm.DB, email string) (models.User, error) {
	var user models.User
	err := db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error
	return user, err
}

func getUserByID(ctx context.Context, db *gorm.DB, id int64) (models.User, error) {
	var user models.User
	err := db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error
	return user, err
}
