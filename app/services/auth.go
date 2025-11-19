package services

import (
	"errors"
	"fmt"
	"time"

	"fiber-starter/app/middleware"
	"fiber-starter/app/models"
	"fiber-starter/config"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService 认证服务接口
type AuthService interface {
	Register(user *models.User) error
	Login(email, password string) (*models.User, string, string, error)
	RefreshToken(refreshToken string) (string, string, error)
	Logout(token string) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	ForgotPassword(email string) error
	ResetPassword(token, email, newPassword string) error
}

// authService 认证服务实现
type authService struct {
	db     *gorm.DB
	config *config.Config
	cache  CacheService
}

// NewAuthService 创建认证服务实例
func NewAuthService(db *gorm.DB, cfg *config.Config, cache CacheService) AuthService {
	return &authService{
		db:     db,
		config: cfg,
		cache:  cache,
	}
}

// Register 用户注册
func (s *authService) Register(user *models.User) error {
	// 检查邮箱是否已存在
	var existingUser models.User
	if err := s.db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		return errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	user.Password = string(hashedPassword)

	// 创建用户
	if err := s.db.Create(user).Error; err != nil {
		return fmt.Errorf("用户创建失败: %w", err)
	}

	return nil
}

// Login 用户登录
func (s *authService) Login(email, password string) (*models.User, string, string, error) {
	var user models.User

	// 查找用户
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", errors.New("邮箱或密码错误")
		}
		return nil, "", "", fmt.Errorf("查询用户失败: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		return nil, "", "", errors.New("用户账户已被禁用")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", "", errors.New("邮箱或密码错误")
	}

	// 生成JWT令牌
	accessToken, err := middleware.GenerateToken(&user, s.config)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成访问令牌失败: %w", err)
	}

	refreshToken, err := middleware.GenerateRefreshToken(&user, s.config)
	if err != nil {
		return nil, "", "", fmt.Errorf("生成刷新令牌失败: %w", err)
	}

	// 将刷新令牌存储到缓存中
	cacheKey := fmt.Sprintf("refresh_token:%d", user.ID)
	if err := s.cache.Set(cacheKey, refreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		// 缓存失败不影响登录，记录日志即可
		fmt.Printf("缓存刷新令牌失败: %v\n", err)
	}

	return &user, accessToken, refreshToken, nil
}

// RefreshToken 刷新访问令牌
func (s *authService) RefreshToken(refreshToken string) (string, string, error) {
	// 验证刷新令牌
	claims, err := middleware.ValidateToken(refreshToken, s.config)
	if err != nil {
		return "", "", errors.New("无效的刷新令牌")
	}

	// 检查缓存中的刷新令牌
	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	cachedToken, err := s.cache.Get(cacheKey)
	if err != nil || cachedToken != refreshToken {
		return "", "", errors.New("刷新令牌已失效")
	}

	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return "", "", fmt.Errorf("用户不存在: %w", err)
	}

	// 检查用户状态
	if !user.IsActive() {
		return "", "", errors.New("用户账户已被禁用")
	}

	// 生成新的令牌
	newAccessToken, err := middleware.GenerateToken(&user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("生成新的访问令牌失败: %w", err)
	}

	newRefreshToken, err := middleware.GenerateRefreshToken(&user, s.config)
	if err != nil {
		return "", "", fmt.Errorf("生成新的刷新令牌失败: %w", err)
	}

	// 更新缓存中的刷新令牌
	if err := s.cache.Set(cacheKey, newRefreshToken, time.Duration(s.config.JWT.RefreshTime)*time.Second); err != nil {
		fmt.Printf("更新刷新令牌缓存失败: %v\n", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// Logout 用户登出
func (s *authService) Logout(token string) error {
	// 验证令牌获取用户ID
	claims, err := middleware.ValidateToken(token, s.config)
	if err != nil {
		return errors.New("无效的令牌")
	}

	// 从缓存中删除刷新令牌
	cacheKey := fmt.Sprintf("refresh_token:%d", claims.UserID)
	if err := s.cache.Delete(cacheKey); err != nil {
		fmt.Printf("删除刷新令牌缓存失败: %v\n", err)
	}

	// 将访问令牌加入黑名单（缓存）
	blacklistKey := fmt.Sprintf("blacklist:%s", token)
	if err := s.cache.Set(blacklistKey, "1", time.Duration(s.config.JWT.ExpirationTime)*time.Second); err != nil {
		fmt.Printf("将令牌加入黑名单失败: %v\n", err)
	}

	return nil
}

// ChangePassword 修改密码
func (s *authService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User

	// 获取用户
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("用户不存在: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("密码更新失败: %w", err)
	}

	return nil
}

// ForgotPassword 忘记密码
func (s *authService) ForgotPassword(email string) error {
	var user models.User

	// 查找用户
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // 为了安全，即使用户不存在也返回成功
		}
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 生成重置令牌
	resetToken := fmt.Sprintf("%d:%d", user.ID, time.Now().Unix())

	// 存储重置令牌到缓存（有效期1小时）
	cacheKey := fmt.Sprintf("reset_token:%s", resetToken)
	if err := s.cache.Set(cacheKey, user.Email, time.Hour); err != nil {
		return fmt.Errorf("存储重置令牌失败: %w", err)
	}

	// TODO: 发送重置密码邮件
	fmt.Printf("重置密码令牌: %s, 邮箱: %s\n", resetToken, user.Email)

	return nil
}

// ResetPassword 重置密码
func (s *authService) ResetPassword(token, email, newPassword string) error {
	// 验证重置令牌
	cacheKey := fmt.Sprintf("reset_token:%s", token)
	cachedEmail, err := s.cache.Get(cacheKey)
	if err != nil || cachedEmail != email {
		return errors.New("无效或已过期的重置令牌")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新用户密码
	if err := s.db.Model(&models.User{}).Where("email = ?", email).Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("密码重置失败: %w", err)
	}

	// 删除重置令牌
	s.cache.Delete(cacheKey)

	return nil
}
