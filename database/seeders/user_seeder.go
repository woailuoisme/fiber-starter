package seeders

import (
	"context"
	"database/sql"

	models "fiber-starter/app/Models"
	"fiber-starter/database/factories"

	"gorm.io/gorm"
)

// UserSeeder 负责用户相关种子数据
type UserSeeder struct{}

// WithUserSeeder 创建用户种子器
func WithUserSeeder() *UserSeeder {
	return &UserSeeder{}
}

// SeedUsers 插入默认用户种子数据
func (s *UserSeeder) SeedUsers(db *sql.DB, dialect string) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		count, err := seedCountUsers(context.Background(), gdb)
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

		return gdb.WithContext(context.Background()).Create(&users).Error
	})
}

// SeedRandomUsers 插入随机用户种子数据
func (s *UserSeeder) SeedRandomUsers(db *sql.DB, dialect string, count int) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		existingCount, err := seedCountUsers(context.Background(), gdb)
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

		return gdb.WithContext(context.Background()).Create(&users).Error
	})
}

// ClearUsers 清空用户种子数据
func (s *UserSeeder) ClearUsers(db *sql.DB, dialect string) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		return gdb.WithContext(context.Background()).
			Where("1 = 1").
			Delete(&models.User{}).Error
	})
}

// CreateAdminUser 创建指定管理员用户
func (s *UserSeeder) CreateAdminUser(db *sql.DB, dialect string, name, email, password string) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		exists, err := seedUserExistsByEmail(context.Background(), gdb, email)
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

		return gdb.WithContext(context.Background()).Create(&admin).Error
	})
}

// GenerateTestUsers 生成测试用户
func (s *UserSeeder) GenerateTestUsers(db *sql.DB, dialect string, count int) error {
	return s.SeedRandomUsers(db, dialect, count)
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
