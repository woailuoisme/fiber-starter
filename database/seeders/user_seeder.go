package seeders

import (
	"context"
	"database/sql"

	models "fiber-starter/app/Models"
	"fiber-starter/database/factories"
	dbsqlc "fiber-starter/database/sqlc"
)

// UserSeeder 负责用户相关种子数据
type UserSeeder struct{}

// WithUserSeeder 创建用户种子器
func WithUserSeeder() *UserSeeder {
	return &UserSeeder{}
}

// SeedUsers 插入默认用户种子数据
func (s *UserSeeder) SeedUsers(db *sql.DB, dialect string) error {
	return withSeedQueries(db, dialect, func(q *dbsqlc.Queries) error {
		count, err := q.CountUsers(context.Background())
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

		return createSeedUsers(context.Background(), q, users)
	})
}

// SeedRandomUsers 插入随机用户种子数据
func (s *UserSeeder) SeedRandomUsers(db *sql.DB, dialect string, count int) error {
	return withSeedQueries(db, dialect, func(q *dbsqlc.Queries) error {
		existingCount, err := q.CountUsers(context.Background())
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

		return createSeedUsers(context.Background(), q, users)
	})
}

// ClearUsers 清空用户种子数据
func (s *UserSeeder) ClearUsers(db *sql.DB, dialect string) error {
	return withSeedQueries(db, dialect, func(q *dbsqlc.Queries) error {
		return q.DeleteAllUsers(context.Background())
	})
}

// CreateAdminUser 创建指定管理员用户
func (s *UserSeeder) CreateAdminUser(db *sql.DB, dialect string, name, email, password string) error {
	return withSeedQueries(db, dialect, func(q *dbsqlc.Queries) error {
		exists, err := q.UserExistsByEmail(context.Background(), email)
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

		_, err = q.CreateUser(context.Background(), dbsqlc.CreateUserParams{
			Name:            admin.Name,
			Email:           admin.Email,
			Password:        admin.Password,
			Avatar:          admin.Avatar,
			Phone:           admin.Phone,
			Status:          admin.Status,
			EmailVerifiedAt: admin.EmailVerifiedAt,
			CreatedAt:       admin.CreatedAt,
			UpdatedAt:       admin.UpdatedAt,
		})
		return err
	})
}

// GenerateTestUsers 生成测试用户
func (s *UserSeeder) GenerateTestUsers(db *sql.DB, dialect string, count int) error {
	return s.SeedRandomUsers(db, dialect, count)
}

func createSeedUsers(ctx context.Context, q *dbsqlc.Queries, users []models.User) error {
	for i := range users {
		user := users[i]
		_, err := q.CreateUser(ctx, dbsqlc.CreateUserParams{
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
			return err
		}
	}
	return nil
}

// Package-level compatibility helpers
func SeedUsers(db *sql.DB, dialect string) error {
	return WithUserSeeder().SeedUsers(db, dialect)
}

func SeedRandomUsers(db *sql.DB, dialect string, count int) error {
	return WithUserSeeder().SeedRandomUsers(db, dialect, count)
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
