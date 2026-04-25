package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	models "fiber-starter/app/Models"
	helpers "fiber-starter/app/Support"
	database "fiber-starter/database"

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

type userService struct {
	db *database.Connection
}

func NewUserService(db *database.Connection) UserService {
	return &userService{db: db}
}

func (s *userService) GetUserByID(id int64) (*models.User, error) {
	var user models.User
	err := withGormDB(s.db, func(db *gorm.DB) error {
		var err error
		user, err = getUserByID(context.Background(), db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			helpers.LogError("Failed to query user", zap.Error(err), zap.Int64("id", id))
			return fmt.Errorf("failed to query user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := withGormDB(s.db, func(db *gorm.DB) error {
		var err error
		user, err = getUserByEmail(context.Background(), db, email)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("user not found")
			}
			helpers.LogError("Failed to query user", zap.Error(err), zap.String("email", email))
			return fmt.Errorf("failed to query user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *userService) GetUsers(page, limit int) ([]models.User, int64, error) {
	return s.listUsers("", page, limit)
}

func (s *userService) UpdateUser(id int64, updates map[string]interface{}) error {
	filtered := userAllowedUpdates(updates)
	if len(filtered) == 0 {
		return nil
	}
	filtered["updated_at"] = utcNow()

	return withGormDB(s.db, func(db *gorm.DB) error {
		err := db.WithContext(context.Background()).
			Model(&models.User{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(filtered).Error
		if err != nil {
			helpers.LogError("Failed to update user", zap.Error(err), zap.Int64("id", id))
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

func (s *userService) DeleteUser(id int64) error {
	now := utcNow()
	return withGormDB(s.db, func(db *gorm.DB) error {
		err := db.WithContext(context.Background()).
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
	})
}

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

func (s *userService) SearchUsers(query string, page, limit int) ([]models.User, int64, error) {
	return s.listUsers(strings.TrimSpace(query), page, limit)
}

func (s *userService) listUsers(search string, page, limit int) ([]models.User, int64, error) {
	_, limit, offset := normalizePagination(page, limit)

	var users []models.User
	var total int64
	err := withGormDB(s.db, func(db *gorm.DB) error {
		query := db.WithContext(context.Background()).
			Model(&models.User{}).
			Where("deleted_at IS NULL")
		if search != "" {
			like := "%" + search + "%"
			query = query.Where("name LIKE ? OR email LIKE ?", like, like)
		}

		if err := query.Count(&total).Error; err != nil {
			helpers.LogError("Failed to get user count", zap.Error(err), zap.String("query", search))
			return fmt.Errorf("failed to get user count: %w", err)
		}

		if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
			helpers.LogError("Failed to get user list", zap.Error(err), zap.String("query", search))
			return fmt.Errorf("failed to get user list: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
