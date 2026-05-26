package seeders

import (
	"context"
	"database/sql"

	"lfiber/database/factories"
	userPkg "lfiber/internal/features/user"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
)

// UserSeeder 负责用户相关种子数据
type UserSeeder struct{}

// WithUserSeeder 创建用户种子器
func WithUserSeeder() *UserSeeder {
	return &UserSeeder{}
}

// SeedUsers 插入默认用户种子数据
func (s *UserSeeder) SeedUsers(db *sql.DB, dialect string) error {
	ctx := context.Background()
	repo := s.repo(db, dialect)

	count, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	factory := factories.NewUserFactory()
	users, err := factory.SeedUsers("password123")
	if err != nil {
		return err
	}

	return s.createSeedUsers(ctx, repo, users)
}

// SeedRandomUsers 插入随机用户种子数据
func (s *UserSeeder) SeedRandomUsers(db *sql.DB, dialect string, count int) error {
	ctx := context.Background()
	repo := s.repo(db, dialect)

	existingCount, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if existingCount >= int64(count) {
		return nil
	}

	factory := factories.NewUserFactory()
	users, err := factory.RandomUsers(count, "password123")
	if err != nil {
		return err
	}

	return s.createSeedUsers(ctx, repo, users)
}

// ClearUsers 清空用户种子数据
func (s *UserSeeder) ClearUsers(db *sql.DB, dialect string) error {
	ctx := context.Background()
	repo := s.repo(db, dialect)
	return repo.DeleteAll(ctx)
}

// CreateAdminUser 创建指定管理员用户
func (s *UserSeeder) CreateAdminUser(db *sql.DB, dialect string, name, email, password string) error {
	ctx := context.Background()
	repo := s.repo(db, dialect)

	exists, err := repo.ExistsByEmail(ctx, email)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	factory := factories.NewUserFactory()
	admin, err := factory.Admin(name, email, password)
	if err != nil {
		return err
	}

	return repo.Create(ctx, &admin)
}

// GenerateTestUsers 生成测试用户
func (s *UserSeeder) GenerateTestUsers(db *sql.DB, dialect string, count int) error {
	return s.SeedRandomUsers(db, dialect, count)
}

func (s *UserSeeder) createSeedUsers(ctx context.Context, repo *userPkg.UserRepository, users []userPkg.User) error {
	for i := range users {
		user := users[i]
		if err := repo.Create(ctx, &user); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserSeeder) repo(db *sql.DB, dialect string) *userPkg.UserRepository {
	var bunDB *bun.DB
	switch dialect {
	case "sqlite":
		bunDB = bun.NewDB(db, sqlitedialect.New())
	default:
		bunDB = bun.NewDB(db, pgdialect.New())
	}
	return userPkg.NewUserRepository(bunDB)
}

// Package-level compatibility helpers
func SeedUsers(db *sql.DB, dialect string) error {
	return WithUserSeeder().SeedUsers(db, dialect)
}

func SeedRandomUsers(db *sql.DB, dialect string, userCount int) error {
	return WithUserSeeder().SeedRandomUsers(db, dialect, userCount)
}

func ClearUsers(db *sql.DB, dialect string) error {
	return WithUserSeeder().ClearUsers(db, dialect)
}

func CreateAdminUser(db *sql.DB, dialect string, name, email, password string) error {
	return WithUserSeeder().CreateAdminUser(db, dialect, name, email, password)
}

func GenerateTestUsers(db *sql.DB, dialect string, count int) error {
	return WithUserSeeder().GenerateTestUsers(db, dialect, count)
}
