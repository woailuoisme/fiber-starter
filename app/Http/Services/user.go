package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	models "fiber-starter/app/Models"
	helpers "fiber-starter/app/Support"
	database "fiber-starter/database"
	dbsqlc "fiber-starter/database/sqlc"

	"go.uber.org/zap"
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
	err := withQueries(s.db, func(q *dbsqlc.Queries) error {
		var err error
		user, err = q.GetUserByID(context.Background(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
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
	err := withQueries(s.db, func(q *dbsqlc.Queries) error {
		var err error
		user, err = q.GetUserByEmail(context.Background(), email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
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

	return withQueries(s.db, func(q *dbsqlc.Queries) error {
		current, err := q.GetUserByID(context.Background(), id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
			helpers.LogError("Failed to query user", zap.Error(err), zap.Int64("id", id))
			return fmt.Errorf("failed to query user: %w", err)
		}

		applyUserUpdates(&current, filtered)
		current.UpdatedAt = utcNow()

		if err := q.UpdateUser(context.Background(), dbsqlc.UpdateUserParams{
			ID:              current.ID,
			Name:            current.Name,
			Email:           current.Email,
			Password:        current.Password,
			Avatar:          current.Avatar,
			Phone:           current.Phone,
			Status:          current.Status,
			EmailVerifiedAt: current.EmailVerifiedAt,
			UpdatedAt:       current.UpdatedAt,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
			helpers.LogError("Failed to update user", zap.Error(err), zap.Int64("id", id))
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

func (s *userService) DeleteUser(id int64) error {
	now := utcNow()
	return withQueries(s.db, func(q *dbsqlc.Queries) error {
		if err := q.SoftDeleteUser(context.Background(), dbsqlc.SoftDeleteUserParams{
			ID:        id,
			DeletedAt: now,
			UpdatedAt: now,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
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
	err := withQueries(s.db, func(q *dbsqlc.Queries) error {
		var err error
		total, err = q.CountUsersBySearch(context.Background(), search)
		if err != nil {
			helpers.LogError("Failed to get user count", zap.Error(err), zap.String("query", search))
			return fmt.Errorf("failed to get user count: %w", err)
		}

		users, err = q.ListUsers(context.Background(), search, limit, offset)
		if err != nil {
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

func applyUserUpdates(user *models.User, updates map[string]interface{}) {
	for key, value := range updates {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "name":
			if str, ok := value.(string); ok {
				user.Name = str
			}
		case "email":
			if str, ok := value.(string); ok {
				user.Email = str
			}
		case "avatar":
			user.Avatar = stringPtrValue(value)
		case "phone":
			user.Phone = stringPtrValue(value)
		case "status":
			if status, ok := value.(models.UserStatus); ok {
				user.Status = status
			} else if str, ok := value.(string); ok {
				user.Status = models.UserStatus(str)
			}
		case "email_verified_at":
			user.EmailVerifiedAt = timePtrValue(value)
		}
	}
}

func stringPtrValue(value interface{}) *string {
	switch v := value.(type) {
	case *string:
		return v
	case string:
		if v == "" {
			return nil
		}
		return &v
	default:
		return nil
	}
}

func timePtrValue(value interface{}) *time.Time {
	switch v := value.(type) {
	case *time.Time:
		return v
	case time.Time:
		return &v
	default:
		return nil
	}
}
