package factories

import (
	"math/rand/v2"
	"time"

	models "fiber-starter/app/Models"

	"github.com/go-faker/faker/v4"
	"golang.org/x/crypto/bcrypt"
)

// UserFactory 提供用户模型的工厂数据
type UserFactory struct{}

// NewUserFactory 创建用户工厂
func NewUserFactory() *UserFactory {
	return &UserFactory{}
}

// Make 生成基础用户数据
func (f *UserFactory) Make(name, email, password string) (models.User, error) {
	return f.baseUser(name, email, password)
}

// Active 标记为已激活状态
func (f *UserFactory) Active(user models.User) models.User {
	user.Status = models.UserStatusActive
	return user
}

// Pending 标记为待激活状态
func (f *UserFactory) Pending(user models.User) models.User {
	user.Status = models.UserStatusInactive
	return user
}

// Banned 标记为封禁状态
func (f *UserFactory) Banned(user models.User) models.User {
	user.Status = models.UserStatusBanned
	return user
}

// Verified 标记邮箱已验证
func (f *UserFactory) Verified(user models.User) models.User {
	user.EmailVerifiedAt = newTimePtr(time.Now().UTC())
	return user
}

// WithPhone 设置手机号
func (f *UserFactory) WithPhone(user models.User, phone string) models.User {
	user.Phone = &phone
	return user
}

// WithAvatar 设置头像
func (f *UserFactory) WithAvatar(user models.User, avatar string) models.User {
	user.Avatar = &avatar
	return user
}

// Admin 生成管理员用户
func (f *UserFactory) Admin(name, email, password string) (models.User, error) {
	user, err := f.Make(name, email, password)
	if err != nil {
		return models.User{}, err
	}

	return f.Verified(f.Active(user)), nil
}

// PendingUser 生成待激活用户
func (f *UserFactory) PendingUser(name, email, password string) (models.User, error) {
	user, err := f.Make(name, email, password)
	if err != nil {
		return models.User{}, err
	}

	return f.Pending(user), nil
}

// BannedUser 生成封禁用户
func (f *UserFactory) BannedUser(name, email, password string) (models.User, error) {
	user, err := f.Make(name, email, password)
	if err != nil {
		return models.User{}, err
	}

	return f.Banned(user), nil
}

// SeedUsers 生成默认种子用户
func (f *UserFactory) SeedUsers(password string) ([]models.User, error) {
	admin, err := f.Admin("Admin User", "admin@example.com", password)
	if err != nil {
		return nil, err
	}

	testUser, err := f.Admin("Test User", "user@example.com", password)
	if err != nil {
		return nil, err
	}

	pendingUser, err := f.PendingUser("Pending User", "pending@example.com", password)
	if err != nil {
		return nil, err
	}

	bannedUser, err := f.BannedUser("Banned User", "banned@example.com", password)
	if err != nil {
		return nil, err
	}

	admin = f.WithPhone(admin, "13800138000")
	testUser = f.WithPhone(testUser, "13800138001")
	pendingUser = f.WithPhone(pendingUser, "13800138002")
	bannedUser = f.WithPhone(bannedUser, "13800138003")

	return []models.User{admin, testUser, pendingUser, bannedUser}, nil
}

// RandomUsers 生成随机用户
func (f *UserFactory) RandomUsers(count int, password string) ([]models.User, error) {
	users := make([]models.User, 0, count)
	now := time.Now().UTC()

	for i := 0; i < count; i++ {
		user, err := f.Make(faker.Name(), faker.Email(), password)
		if err != nil {
			return nil, err
		}

		user = f.WithPhone(user, faker.Phonenumber())
		user.CreatedAt = now
		user.UpdatedAt = now

		hasVerifiedAt := randomInt(3)
		if hasVerifiedAt == 0 {
			daysAgo := randomInt(30)
			daysAgo++
			verifiedAt := now.AddDate(0, 0, -daysAgo)
			user = f.Active(user)
			user.EmailVerifiedAt = newTimePtr(verifiedAt)
		}

		isPending := randomInt(10)
		if isPending == 0 {
			user = f.Pending(user)
		}

		users = append(users, user)
	}

	return users, nil
}

func (f *UserFactory) baseUser(name, email, password string) (models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, err
	}

	now := time.Now().UTC()
	return models.User{
		Name:      name,
		Email:     email,
		Password:  string(hashedPassword),
		Status:    models.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func newTimePtr(value time.Time) *time.Time {
	ptr := new(time.Time)
	*ptr = value
	return ptr
}

//nolint:gosec // factory 仅用于测试/种子数据的随机生成
func randomInt(max int) int {
	if max <= 0 {
		return 0
	}

	return rand.IntN(max)
}
