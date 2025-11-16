package services

import (
	"errors"
	"fmt"

	"fiber-starter/app/models"

	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUsers(page, limit int) ([]models.User, int64, error)
	UpdateUser(id uint, updates map[string]interface{}) error
	DeleteUser(id uint) error
	UpdateProfile(id uint, profile *models.User) error
	SearchUsers(query string, page, limit int) ([]models.User, int64, error)
}

// userService 用户服务实现
type userService struct {
	db *gorm.DB
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB) UserService {
	return &userService{
		db: db,
	}
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// GetUsers 获取用户列表（分页）
func (s *userService) GetUsers(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取总数
	if err := s.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 获取分页数据
	if err := s.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("获取用户列表失败: %w", err)
	}

	return users, total, nil
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(id uint, updates map[string]interface{}) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 如果更新邮箱，检查是否已存在
	if email, ok := updates["email"].(string); ok && email != user.Email {
		var existingUser models.User
		if err := s.db.Where("email = ? AND id != ?", email, id).First(&existingUser).Error; err == nil {
			return errors.New("邮箱已被其他用户使用")
		}
	}

	// 更新用户信息
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}

	return nil
}

// DeleteUser 删除用户（软删除）
func (s *userService) DeleteUser(id uint) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 软删除用户
	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// UpdateProfile 更新用户资料
func (s *userService) UpdateProfile(id uint, profile *models.User) error {
	// 检查用户是否存在
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 更新允许的字段
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

	// 执行更新
	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新用户资料失败: %w", err)
	}

	return nil
}

// SearchUsers 搜索用户
func (s *userService) SearchUsers(query string, page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// 计算偏移量
	offset := (page - 1) * limit

	// 构建搜索条件
	searchPattern := "%" + query + "%"
	
	// 获取总数
	if err := s.db.Model(&models.User{}).Where("name LIKE ? OR email LIKE ?", searchPattern, searchPattern).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取搜索结果总数失败: %w", err)
	}

	// 获取搜索结果
	if err := s.db.Where("name LIKE ? OR email LIKE ?", searchPattern, searchPattern).
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("搜索用户失败: %w", err)
	}

	return users, total, nil
}