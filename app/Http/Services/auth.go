package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"fiber-starter/app/Http/Middleware"
	models "fiber-starter/app/Models"
	helpers "fiber-starter/app/Support"
	"fiber-starter/config"
	database "fiber-starter/database"
	dbsqlc "fiber-starter/database/sqlc"

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

type authService struct {
	db     *database.Connection
	config *config.Config
	cache  helpers.CacheService
}

func NewAuthService(db *database.Connection, cfg *config.Config, cache helpers.CacheService) AuthService {
	return &authService{db: db, config: cfg, cache: cache}
}

func (s *authService) Register(user *models.User) error {
	return withQueries(s.db, func(q *dbsqlc.Queries) error {
		exists, err := q.UserExistsByEmail(context.Background(), user.Email)
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

		if user.Status == "" {
			user.Status = models.UserStatusActive
		}
		now := utcNow()
		user.CreatedAt = now
		user.UpdatedAt = now

		_, err = q.CreateUser(context.Background(), dbsqlc.CreateUserParams{
			Name:            user.Name,
			Email:           user.Email,
			Password:        user.Password,
			Avatar:          user.Avatar,
			Phone:           user.Phone,
			Status:          user.Status,
			EmailVerifiedAt: user.EmailVerifiedAt,
			CreatedAt:       user.CreatedAt,
			UpdatedAt:       user.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	})
}

func (s *authService) Login(email, password string) (*models.User, string, string, error) {
	var user models.User
	var accessToken, refreshToken string

	err := withQueries(s.db, func(q *dbsqlc.Queries) error {
		var err error
		user, err = q.GetUserByEmail(context.Background(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("invalid email or password")
			}
			helpers.LogError("Failed to query user", zap.Error(err))
			return fmt.Errorf("failed to query user: %w", err)
		}

		if !user.IsActive() {
			return errors.New("user account has been disabled")
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			return errors.New("invalid email or password")
		}

		accessToken, err = middleware.GenerateToken(&user, s.config)
		if err != nil {
			return fmt.Errorf("failed to generate access token: %w", err)
		}

		refreshToken, err = middleware.GenerateRefreshToken(&user, s.config)
		if err != nil {
			return fmt.Errorf("failed to generate refresh token: %w", err)
		}

		cacheKey := fmt.Sprintf("refresh_token:%d", user.ID)
		if err := s.cache.Set(cacheKey, refreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
			helpers.LogError("Failed to cache refresh token", zap.Error(err))
		}

		return nil
	})
	if err != nil {
		return nil, "", "", err
	}

	return &user, accessToken, refreshToken, nil
}

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

	var user models.User
	var newAccessToken, newRefreshToken string
	err = withQueries(s.db, func(q *dbsqlc.Queries) error {
		var err error
		user, err = q.GetUserByID(context.Background(), claims.UserID)
		if err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		if !user.IsActive() {
			return errors.New("user account has been disabled")
		}

		newAccessToken, err = middleware.GenerateToken(&user, s.config)
		if err != nil {
			return fmt.Errorf("failed to generate new access token: %w", err)
		}

		newRefreshToken, err = middleware.GenerateRefreshToken(&user, s.config)
		if err != nil {
			return fmt.Errorf("failed to generate new refresh token: %w", err)
		}

		if err := s.cache.Set(cacheKey, newRefreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
			helpers.LogError("Failed to update refresh token cache", zap.Error(err))
		}

		return nil
	})
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

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

func (s *authService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	return withQueries(s.db, func(q *dbsqlc.Queries) error {
		user, err := q.GetUserByID(context.Background(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
			return fmt.Errorf("user not found: %w", err)
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
			return errors.New("incorrect current password")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		if err := q.UpdatePassword(context.Background(), dbsqlc.UpdatePasswordParams{
			ID:        userID,
			Password:  string(hashedPassword),
			UpdatedAt: utcNow(),
		}); err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		return nil
	})
}

func (s *authService) ForgotPassword(email string) error {
	return withQueries(s.db, func(q *dbsqlc.Queries) error {
		user, err := q.GetUserByEmail(context.Background(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
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
	})
}

func (s *authService) ResetPassword(token, email, newPassword string) error {
	cacheKey := fmt.Sprintf("reset_token:%s", token)
	cachedEmail, err := s.cache.Get(cacheKey)
	if err != nil || cachedEmail != email {
		return errors.New("invalid or expired reset token")
	}

	return withQueries(s.db, func(q *dbsqlc.Queries) error {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		if err := q.ResetPasswordByEmail(context.Background(), dbsqlc.ResetPasswordByEmailParams{
			Email:     email,
			Password:  string(hashedPassword),
			UpdatedAt: utcNow(),
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
			return fmt.Errorf("failed to reset password: %w", err)
		}

		_ = s.cache.Delete(cacheKey)
		return nil
	})
}
