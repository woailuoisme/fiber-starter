package seeders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	models "fiber-starter/internal/domain/model"

	"github.com/go-faker/faker/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SeedUsers 创建用户种子数据
func SeedUsers(db *sql.DB, dialect string) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		count, err := seedCountUsers(context.Background(), gdb)
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		users := []models.User{
			{Name: "Admin User", Email: "admin@example.com", Password: string(hashedPassword), Phone: stringPtr("13800138000"), Status: models.UserStatusActive, EmailVerifiedAt: &now, CreatedAt: now, UpdatedAt: now},
			{Name: "Test User", Email: "user@example.com", Password: string(hashedPassword), Phone: stringPtr("13800138001"), Status: models.UserStatusActive, EmailVerifiedAt: &now, CreatedAt: now, UpdatedAt: now},
			{Name: "Pending User", Email: "pending@example.com", Password: string(hashedPassword), Phone: stringPtr("13800138002"), Status: models.UserStatusInactive, CreatedAt: now, UpdatedAt: now},
			{Name: "Banned User", Email: "banned@example.com", Password: string(hashedPassword), Phone: stringPtr("13800138003"), Status: models.UserStatusBanned, CreatedAt: now, UpdatedAt: now},
		}

		return gdb.WithContext(context.Background()).Create(&users).Error
	})
}

// SeedRandomUsers 生成随机用户种子数据
func SeedRandomUsers(db *sql.DB, dialect string, count int) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		existingCount, err := seedCountUsers(context.Background(), gdb)
		if err != nil {
			return err
		}
		if existingCount >= int64(count) {
			return nil
		}

		users, err := randomUsers(count, "password123")
		if err != nil {
			return err
		}

		return gdb.WithContext(context.Background()).Create(&users).Error
	})
}

// ClearUsers 清除用户种子数据
func ClearUsers(db *sql.DB, dialect string) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		return gdb.WithContext(context.Background()).
			Where("1 = 1").
			Delete(&models.User{}).Error
	})
}

// CreateAdminUser 创建管理员用户
func CreateAdminUser(db *sql.DB, dialect string, name, email, password string) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		exists, err := seedUserExistsByEmail(context.Background(), gdb, email)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		admin := models.User{
			Name:            name,
			Email:           email,
			Password:        string(hashedPassword),
			Status:          models.UserStatusActive,
			EmailVerifiedAt: &now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}

		return gdb.WithContext(context.Background()).Create(&admin).Error
	})
}

// GenerateTestUsers 生成指定数量的测试用户（使用 faker 和 carbon）
func GenerateTestUsers(db *sql.DB, dialect string, count int) error {
	return withSeedGormDB(db, dialect, func(gdb *gorm.DB) error {
		users, err := randomUsers(count, "test123456")
		if err != nil {
			return err
		}

		return gdb.WithContext(context.Background()).Create(&users).Error
	})
}

func seedCountUsers(ctx context.Context, db *gorm.DB) (int64, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

func seedUserExistsByEmail(ctx context.Context, db *gorm.DB, email string) (bool, error) {
	var count int64
	err := db.WithContext(ctx).Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

func seedGormDB(db *sql.DB, dialect string) (*gorm.DB, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	switch strings.ToLower(strings.TrimSpace(dialect)) {
	case "psql", "postgres", "postgresql":
		return gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	case "sqlite", "sqlite3":
		return gorm.Open(sqlite.Dialector{Conn: db}, &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", dialect)
	}
}

func randomUsers(count int, password string) ([]models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	users := make([]models.User, 0, count)
	now := time.Now().UTC()
	for i := 0; i < count; i++ {
		phone := faker.Phonenumber()
		user := models.User{
			Name:      faker.Name(),
			Email:     faker.Email(),
			Password:  string(hashedPassword),
			Phone:     &phone,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if rand.Intn(3) == 0 { //nolint:gosec // seeder
			daysAgo := rand.Intn(30) + 1 //nolint:gosec // seeder
			verifiedAt := now.AddDate(0, 0, -daysAgo)
			user.EmailVerifiedAt = &verifiedAt
		}

		if rand.Intn(10) == 0 { //nolint:gosec // seeder
			user.Status = models.UserStatusInactive
		}

		users = append(users, user)
	}

	return users, nil
}

func stringPtr(value string) *string {
	return &value
}
