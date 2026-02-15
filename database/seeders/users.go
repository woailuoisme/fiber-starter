package seeders

import (
	"math/rand"
	"time"

	"fiber-starter/app/models"
	"fiber-starter/database"

	"github.com/go-faker/faker/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedUsers 创建用户种子数据
func SeedUsers(db *gorm.DB) error {
	// 检查是否已有用户数据
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	// 如果已有数据，跳过种子数据创建
	if count > 0 {
		return nil
	}

	// 创建测试用户密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 准备种子用户数据
	users := []models.User{
		{
			Name:     "管理员",
			Email:    "admin@example.com",
			Password: string(hashedPassword),
			Phone:    stringPtr("13800138000"),
			Status:   models.UserStatusActive,
		},
		{
			Name:     "测试用户",
			Email:    "user@example.com",
			Password: string(hashedPassword),
			Phone:    stringPtr("13800138001"),
			Status:   models.UserStatusActive,
		},
		{
			Name:     "待验证用户",
			Email:    "pending@example.com",
			Password: string(hashedPassword),
			Phone:    stringPtr("13800138002"),
			Status:   models.UserStatusInactive,
		},
		{
			Name:     "被封禁用户",
			Email:    "banned@example.com",
			Password: string(hashedPassword),
			Phone:    stringPtr("13800138003"),
			Status:   models.UserStatusBanned,
		},
	}

	// 设置邮箱验证时间
	now := time.Now()
	users[0].EmailVerifiedAt = &now
	users[1].EmailVerifiedAt = &now

	// 批量创建用户
	if err := db.CreateInBatches(users, 100).Error; err != nil {
		return err
	}

	return nil
}

// SeedRandomUsers 生成随机用户种子数据
func SeedRandomUsers(db *gorm.DB, count int) error {
	// 检查是否已有用户数据
	var existingCount int64
	if err := db.Model(&models.User{}).Count(&existingCount).Error; err != nil {
		return err
	}

	// 如果已有足够的数据，跳过种子数据创建
	if existingCount >= int64(count) {
		return nil
	}

	// 创建测试用户密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 生成随机用户数据
	users := make([]models.User, 0, count)
	for i := 0; i < count; i++ {
		// 使用 faker 生成随机数据
		name := faker.Name()
		email := faker.Email()
		phone := faker.Phonenumber()

		user := models.User{
			Name:     name,
			Email:    email,
			Password: string(hashedPassword),
			Phone:    &phone,
			Status:   models.UserStatusActive,
		}

		// 随机设置一些用户为已验证邮箱
		if i%3 == 0 { // 每3个用户中有1个已验证邮箱
			daysAgo := rand.Intn(30) + 1 //nolint:gosec // seeder
			verifiedTime := time.Now().AddDate(0, 0, -daysAgo)
			user.EmailVerifiedAt = &verifiedTime
		}

		// 随机设置一些用户为非活跃状态
		if i%10 == 0 { // 每10个用户中有1个为非活跃
			user.Status = models.UserStatusInactive
		}

		users = append(users, user)
	}

	// 批量创建用户
	if err := db.CreateInBatches(users, 100).Error; err != nil {
		return err
	}

	return nil
}

// ClearUsers 清除用户种子数据
func ClearUsers(db *gorm.DB) error {
	// 删除所有用户数据
	return db.Exec("DELETE FROM users").Error
}

// SeedUserTable 执行用户表种子数据
func SeedUserTable() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	// 执行种子数据创建
	if err := SeedUsers(db); err != nil {
		return err
	}

	return nil
}

// ClearUserTable 清除用户表种子数据
func ClearUserTable() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	// 执行数据清除
	if err := ClearUsers(db); err != nil {
		return err
	}

	return nil
}

// stringPtr 返回字符串指针的辅助函数
func stringPtr(s string) *string {
	return &s
}

// CreateAdminUser 创建管理员用户
func CreateAdminUser(db *gorm.DB, name, email, password string) error {
	// 检查管理员是否已存在
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil // 管理员已存在
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 创建管理员用户
	admin := models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Status:   models.UserStatusActive,
	}

	// 设置邮箱验证时间
	now := time.Now()
	admin.EmailVerifiedAt = &now

	return db.Create(&admin).Error
}

// GenerateTestUsers 生成指定数量的测试用户（使用 faker 和 carbon）
func GenerateTestUsers(db *gorm.DB, count int) error {
	// 创建测试用户密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("test123456"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 生成随机用户数据
	users := make([]models.User, 0, count)
	for i := 0; i < count; i++ {
		// 使用 faker 生成更真实的测试数据
		name := faker.Name()
		email := faker.Email()
		phone := faker.Phonenumber()

		user := models.User{
			Name:     name,
			Email:    email,
			Password: string(hashedPassword),
			Phone:    &phone,
			Status:   models.UserStatusActive,
		}

		// 随机设置邮箱验证状态
		if rand.Intn(2) == 1 { //nolint:gosec // seeder
			// 生成随机的验证时间（过去30天内）
			daysAgo := rand.Intn(30) + 1 //nolint:gosec // seeder
			verifiedTime := time.Now().AddDate(0, 0, -daysAgo)
			user.EmailVerifiedAt = &verifiedTime
		}

		// 随机设置用户状态
		user.Status = models.UserStatusActive
		if rand.Intn(10) == 0 { //nolint:gosec // seeder // 10% 概率为非活跃
			user.Status = models.UserStatusInactive
		}

		users = append(users, user)
	}

	// 批量创建用户
	return db.CreateInBatches(users, 100).Error
}
