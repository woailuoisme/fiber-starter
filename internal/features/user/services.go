package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	exceptions "fiber-starter/internal/common/exceptions"
	helpers "fiber-starter/internal/common/support"
	database "fiber-starter/internal/providers/database/Contracts"

	"github.com/uptrace/bun"
	"go.uber.org/zap"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(ctx context.Context, id int64) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListUsers(ctx context.Context, query UserListQuery) (UserPage, error)
	UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (*User, error)
	DeleteUser(ctx context.Context, id int64) error
	UpdateProfile(ctx context.Context, id int64, input UpdateUserInput) (*User, error)
	SearchUsers(ctx context.Context, query UserListQuery) (UserPage, error)
}

type UserListQuery struct {
	Search string
	Page   int
	Limit  int
}

type UserPage struct {
	Items []User
	Total int64
	Page  int
	Limit int
}

type UpdateUserInput struct {
	Name            *string
	Email           *string
	Avatar          *string
	Phone           *string
	Status          *UserStatus
	EmailVerifiedAt *time.Time
}

func (i UpdateUserInput) IsZero() bool {
	return i.Name == nil &&
		i.Email == nil &&
		i.Avatar == nil &&
		i.Phone == nil &&
		i.Status == nil &&
		i.EmailVerifiedAt == nil
}

type userService struct {
	db database.Connection
}

func NewUserService(db database.Connection) UserService {
	return &userService{db: db}
}

func serviceContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (s *userService) userRepo() (*UserRepository, error) {
	if s == nil || s.db == nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", errors.New("database connection not initialized"))
	}
	bunDB, err := s.db.BunDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}
	return NewUserRepository(bunDB), nil
}

func (s *userService) bunDB() (*bun.DB, error) {
	if s == nil || s.db == nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", errors.New("database connection not initialized"))
	}
	bunDB, err := s.db.BunDB()
	if err != nil {
		return nil, exceptions.ServiceUnavailableWithCause("Service unavailable", err)
	}
	return bunDB, nil
}

func (s *userService) GetUserByID(ctx context.Context, id int64) (*User, error) {
	ctx = serviceContext(ctx)
	repo, err := s.userRepo()
	if err != nil {
		return nil, err
	}

	user, err := repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		helpers.LogError("Failed to query user", zap.Error(err), zap.Int64("id", id))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	ctx = serviceContext(ctx)
	repo, err := s.userRepo()
	if err != nil {
		return nil, err
	}

	user, err := repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		helpers.LogError("Failed to query user", zap.Error(err), zap.String("email", email))
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

func (s *userService) ListUsers(ctx context.Context, query UserListQuery) (UserPage, error) {
	query.Search = ""
	return s.listUsers(ctx, query)
}

func (s *userService) UpdateUser(ctx context.Context, id int64, input UpdateUserInput) (*User, error) {
	ctx = serviceContext(ctx)
	if input.IsZero() {
		return s.GetUserByID(ctx, id)
	}

	bunDB, err := s.bunDB()
	if err != nil {
		return nil, err
	}

	var updated *User
	err = bunDB.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		repo := NewUserRepository(tx)

		current, err := repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("user not found")
			}
			helpers.LogError("Failed to query user", zap.Error(err), zap.Int64("id", id))
			return fmt.Errorf("failed to query user: %w", err)
		}

		applyUserInput(current, input)
		current.UpdatedAt = helpers.UtcNow()

		if err := repo.Update(ctx, current); err != nil {
			helpers.LogError("Failed to update user", zap.Error(err), zap.Int64("id", id))
			return fmt.Errorf("failed to update user: %w", err)
		}

		updated = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	ctx = serviceContext(ctx)
	repo, err := s.userRepo()
	if err != nil {
		return err
	}
	now := helpers.UtcNow()

	if err := repo.SoftDelete(ctx, id, now, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("user not found")
		}
		helpers.LogError("Failed to delete user", zap.Error(err), zap.Int64("id", id))
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func (s *userService) UpdateProfile(ctx context.Context, id int64, input UpdateUserInput) (*User, error) {
	return s.UpdateUser(ctx, id, input)
}

func (s *userService) SearchUsers(ctx context.Context, query UserListQuery) (UserPage, error) {
	return s.listUsers(ctx, query)
}

func (s *userService) listUsers(ctx context.Context, query UserListQuery) (UserPage, error) {
	ctx = serviceContext(ctx)
	page, limit, offset := helpers.NormalizePagination(query.Page, query.Limit)
	search := strings.TrimSpace(query.Search)
	repo, err := s.userRepo()
	if err != nil {
		return UserPage{}, err
	}

	total, err := repo.CountBySearch(ctx, search)
	if err != nil {
		helpers.LogError("Failed to get user count", zap.Error(err), zap.String("query", search))
		return UserPage{}, fmt.Errorf("failed to get user count: %w", err)
	}

	users, err := repo.List(ctx, search, limit, offset)
	if err != nil {
		helpers.LogError("Failed to get user list", zap.Error(err), zap.String("query", search))
		return UserPage{}, fmt.Errorf("failed to get user list: %w", err)
	}

	return UserPage{Items: users, Total: total, Page: page, Limit: limit}, nil
}

func applyUserInput(user *User, input UpdateUserInput) {
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Avatar != nil {
		user.Avatar = input.Avatar
	}
	if input.Phone != nil {
		user.Phone = input.Phone
	}
	if input.Status != nil {
		user.Status = *input.Status
	}
	if input.EmailVerifiedAt != nil {
		user.EmailVerifiedAt = input.EmailVerifiedAt
	}
}
