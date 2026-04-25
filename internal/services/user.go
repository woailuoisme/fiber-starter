package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	database "fiber-starter/internal/db"
	models "fiber-starter/internal/domain/model"
	"fiber-starter/internal/platform/helpers"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(id int64) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUsers(page, limit int) ([]models.User, int64, error)
	UpdateUser(id int64, updates map[string]interface{}) error
	DeleteUser(id int64) error
	UpdateProfile(id int64, profile *models.User) error
	SearchUsers(query string, page, limit int) ([]models.User, int64, error)
}

// userService 用户服务实现
type userService struct {
	db *database.Connection
}

// NewUserService 创建用户服务实例
func NewUserService(db *database.Connection) UserService {
	return &userService{
		db: db,
	}
}

// GetUserByID Get user by ID
func (s *userService) GetUserByID(id int64) (*models.User, error) {
	db, err := s.db.GetGormDB()
	if err != nil {
		return nil, err
	}

	var user models.User
	err = db.WithContext(context.Background()).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		helpers.LogError("Failed to query user", zap.Error(err), zap.Int64("id", id))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail Get user by email
func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	db, err := s.db.GetGormDB()
	if err != nil {
		return nil, err
	}

	var user models.User
	err = db.WithContext(context.Background()).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		helpers.LogError("Failed to query user", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return &user, nil
}

// GetUsers Get user list (paginated)
func (s *userService) GetUsers(page, limit int) ([]models.User, int64, error) {
	return s.listUsers("", page, limit)
}

// UpdateUser Update user information
func (s *userService) UpdateUser(id int64, updates map[string]interface{}) error {
	db, err := s.db.GetGormDB()
	if err != nil {
		return err
	}

	filtered := userAllowedUpdates(updates)
	if len(filtered) == 0 {
		return nil
	}
	filtered["updated_at"] = time.Now().UTC()

	err = db.WithContext(context.Background()).
		Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(filtered).Error
	if err != nil {
		helpers.LogError("Failed to update user", zap.Error(err), zap.Int64("id", id))
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser Delete user (soft delete)
func (s *userService) DeleteUser(id int64) error {
	db, err := s.db.GetGormDB()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	err = db.WithContext(context.Background()).
		Model(&models.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"updated_at": now,
		}).Error
	if err != nil {
		helpers.LogError("Failed to delete user", zap.Error(err), zap.Int64("id", id))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// UpdateProfile Update user profile
func (s *userService) UpdateProfile(id int64, profile *models.User) error {
	if profile == nil {
		return nil
	}

	updates := make(map[string]interface{})
	if profile.Name != "" {
		updates["name"] = profile.Name
	}
	if profile.Avatar != nil {
		updates["avatar"] = profile.Avatar
	}
	if profile.Phone != nil {
		updates["phone"] = profile.Phone
	}

	return s.UpdateUser(id, updates)
}

// SearchUsers Search users
func (s *userService) SearchUsers(query string, page, limit int) ([]models.User, int64, error) {
	return s.listUsers(strings.TrimSpace(query), page, limit)
}

func (s *userService) listUsers(search string, page, limit int) ([]models.User, int64, error) {
	db, err := s.db.GetGormDB()
	if err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	query := db.WithContext(context.Background()).
		Model(&models.User{}).
		Where("deleted_at IS NULL")
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		helpers.LogError("Failed to get user count", zap.Error(err), zap.String("query", search))
		return nil, 0, fmt.Errorf("failed to get user count: %w", err)
	}

	var users []models.User
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		helpers.LogError("Failed to get user list", zap.Error(err), zap.String("query", search))
		return nil, 0, fmt.Errorf("failed to get user list: %w", err)
	}

	return users, total, nil
}

func userAllowedUpdates(updates map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	for k, v := range updates {
		field := strings.ToLower(strings.TrimSpace(k))
		switch field {
		case "name", "email", "avatar", "phone", "status", "email_verified_at":
			filtered[field] = v
		default:
		}
	}
	return filtered
}
