package seeders

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"strings"
	"time"

	models "fiber-starter/internal/domain/model"

	"github.com/go-faker/faker/v4"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/scan"
	"golang.org/x/crypto/bcrypt"
)

// SeedUsers 创建用户种子数据
func SeedUsers(db *sql.DB, dialect string) error {
	ctx := context.Background()
	if db == nil {
		return errors.New("db is nil")
	}

	count, err := seedCountUsers(ctx, db, dialect)
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

	for _, u := range users {
		q, err := seedInsertUserQuery(dialect, u)
		if err != nil {
			return err
		}
		if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
			return err
		}
	}

	return nil
}

// SeedRandomUsers 生成随机用户种子数据
func SeedRandomUsers(db *sql.DB, dialect string, count int) error {
	ctx := context.Background()
	if db == nil {
		return errors.New("db is nil")
	}

	existingCount, err := seedCountUsers(ctx, db, dialect)
	if err != nil {
		return err
	}
	if existingCount >= int64(count) {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		now := time.Now().UTC()
		name := faker.Name()
		email := faker.Email()
		phone := faker.Phonenumber()

		user := models.User{
			Name:      name,
			Email:     email,
			Password:  string(hashedPassword),
			Phone:     &phone,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if i%3 == 0 { //nolint:gosec // seeder
			daysAgo := rand.Intn(30) + 1 //nolint:gosec // seeder
			verifiedTime := now.AddDate(0, 0, -daysAgo)
			user.EmailVerifiedAt = &verifiedTime
		}

		if i%10 == 0 { //nolint:gosec // seeder
			user.Status = models.UserStatusInactive
		}

		q, err := seedInsertUserQuery(dialect, user)
		if err != nil {
			return err
		}
		if _, err := bob.Exec(ctx, bob.NewDB(db), q); err != nil {
			return err
		}
	}

	return nil
}

// ClearUsers 清除用户种子数据
func ClearUsers(db *sql.DB, dialect string) error {
	ctx := context.Background()
	if db == nil {
		return errors.New("db is nil")
	}

	q, err := seedRawQuery(dialect, "DELETE FROM users")
	if err != nil {
		return err
	}
	_, err = bob.Exec(ctx, bob.NewDB(db), q)
	return err
}

// stringPtr 返回字符串指针的辅助函数
func stringPtr(s string) *string {
	return &s
}

// CreateAdminUser 创建管理员用户
func CreateAdminUser(db *sql.DB, dialect string, name, email, password string) error {
	ctx := context.Background()
	if db == nil {
		return errors.New("db is nil")
	}

	exists, err := seedUserExistsByEmail(ctx, db, dialect, email)
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

	q, err := seedInsertUserQuery(dialect, admin)
	if err != nil {
		return err
	}

	_, err = bob.Exec(ctx, bob.NewDB(db), q)
	return err
}

// GenerateTestUsers 生成指定数量的测试用户（使用 faker 和 carbon）
func GenerateTestUsers(db *sql.DB, dialect string, count int) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		now := time.Now().UTC()
		name := faker.Name()
		email := faker.Email()
		phone := faker.Phonenumber()

		user := models.User{
			Name:      name,
			Email:     email,
			Password:  string(hashedPassword),
			Phone:     &phone,
			Status:    models.UserStatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if rand.Intn(2) == 1 { //nolint:gosec // seeder
			daysAgo := rand.Intn(30) + 1 //nolint:gosec // seeder
			verifiedTime := now.AddDate(0, 0, -daysAgo)
			user.EmailVerifiedAt = &verifiedTime
		}

		if rand.Intn(10) == 0 { //nolint:gosec // seeder
			user.Status = models.UserStatusInactive
		}

		q, err := seedInsertUserQuery(dialect, user)
		if err != nil {
			return err
		}
		if _, err := bob.Exec(context.Background(), bob.NewDB(db), q); err != nil {
			return err
		}
	}

	return nil
}

func seedCountUsers(ctx context.Context, db *sql.DB, dialect string) (int64, error) {
	q, err := seedRawQuery(dialect, "SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, err
	}
	return bob.One(ctx, bob.NewDB(db), q, scan.SingleColumnMapper[int64])
}

func seedUserExistsByEmail(ctx context.Context, db *sql.DB, dialect string, email string) (bool, error) {
	sqlStr := "SELECT 1 FROM users WHERE email = ? LIMIT 1"
	q, err := seedRawQuery(dialect, sqlStr, email)
	if err != nil {
		return false, err
	}

	_, err = bob.One(ctx, bob.NewDB(db), q, scan.SingleColumnMapper[int])
	if err == nil {
		return true, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return false, err
}

func seedInsertUserQuery(dialect string, user models.User) (bob.Query, error) {
	sqlStr := strings.Join([]string{
		"INSERT INTO users (name, email, password, avatar, phone, status, email_verified_at, created_at, updated_at, deleted_at)",
		"VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
	}, " ")
	return seedRawQuery(dialect, sqlStr,
		user.Name,
		user.Email,
		user.Password,
		user.Avatar,
		user.Phone,
		string(user.Status),
		user.EmailVerifiedAt,
		user.CreatedAt,
		user.UpdatedAt,
		user.DeletedAt,
	)
}

func seedRawQuery(dialect string, q string, args ...any) (bob.Query, error) {
	switch dialect {
	case "psql":
		return psql.RawQuery(q, args...), nil
	case "sqlite":
		return sqlite.RawQuery(q, args...), nil
	default:
		return nil, errors.New("unsupported dialect")
	}
}
