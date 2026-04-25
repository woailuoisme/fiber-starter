package middleware

import (
	"fmt"
	"strings"
	"time"

	"fiber-starter/internal/config"
	models "fiber-starter/internal/domain/model"
	"fiber-starter/internal/platform/exceptions"
	"fiber-starter/internal/platform/helpers"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构体
type JWTClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// BearerSchema Bearer 认证前缀
const BearerSchema = "Bearer"

func parseBearerToken(authHeader string) (token string, present bool, validFormat bool) {
	if authHeader == "" {
		return "", false, false
	}

	present = true
	scheme, token, ok := strings.Cut(authHeader, " ")
	if !ok || scheme != BearerSchema || token == "" {
		return "", present, false
	}

	return token, present, true
}

func setUserContext(c fiber.Ctx, claims *JWTClaims) {
	user := &models.User{
		ID:    claims.UserID,
		Email: claims.Email,
		Name:  claims.Name,
	}

	c.Locals("user", user)
	c.Locals("user_id", claims.UserID)
	c.Locals("user_email", claims.Email)
	c.Locals("user_name", claims.Name)
	c.Locals("user_claims", claims)
}

// JWTAuth JWT authentication middleware
func JWTAuth(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenString, present, validFormat := parseBearerToken(c.Get("Authorization"))
		if !present {
			return exceptions.NewAuthenticationException("Missing Authorization header")
		}
		if !validFormat {
			return exceptions.NewAuthenticationException("Invalid Authorization format")
		}

		// Parse and verify token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(_ *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			return exceptions.NewAuthenticationException("Invalid token")
		}

		// Verify token validity
		if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
			setUserContext(c, claims)
			return c.Next()
		}

		return exceptions.NewAuthenticationException("Invalid token")
	}
}

// OptionalJWTAuth 可选JWT认证中间件（不强制要求认证）
func OptionalJWTAuth(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		tokenString, present, validFormat := parseBearerToken(c.Get("Authorization"))
		if !present || !validFormat {
			return c.Next()
		}

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(_ *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT.Secret), nil
		})

		if err == nil {
			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				setUserContext(c, claims)
			}
		}

		return c.Next()
	}
}

// JWTProtected JWT middleware to protect routes (optional: check logout blacklist)
func JWTProtected(cfg *config.Config, cache helpers.CacheService) fiber.Handler {
	return func(c fiber.Ctx) error {
		if cfg == nil {
			return exceptions.NewAPIException("Config not initialized", fiber.StatusInternalServerError)
		}

		token := GetTokenFromContext(c)
		if token == "" {
			return exceptions.NewAuthenticationException("Missing or invalid Authorization")
		}

		if cache != nil {
			blacklistKey := fmt.Sprintf("blacklist:%s", token)
			exists, err := cache.Exists(blacklistKey)
			if err != nil {
				return exceptions.NewAPIException("Auth service unavailable", fiber.StatusServiceUnavailable)
			}
			if exists {
				return exceptions.NewAuthenticationException("Token has been invalidated")
			}
		}

		return JWTAuth(cfg)(c)
	}
}

// GenerateToken 生成JWT令牌
func GenerateToken(user *models.User, cfg *config.Config) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.ExpirationTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "user_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateRefreshToken 生成刷新令牌
func GenerateRefreshToken(user *models.User, cfg *config.Config) (string, error) {
	claims := JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.RefreshTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.JWT.Issuer,
			Subject:   "refresh_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ValidateToken 验证JWT令牌
func ValidateToken(tokenString string, cfg *config.Config) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// GetUserFromContext 从Fiber上下文中获取用户信息
func GetUserFromContext(c fiber.Ctx) *models.User {
	user, ok := c.Locals("user").(*models.User)
	if !ok {
		return nil
	}
	return user
}

// GetCurrentUser 从上下文获取当前用户信息
func GetCurrentUser(c fiber.Ctx) *JWTClaims {
	if claims, ok := c.Locals("user_claims").(*JWTClaims); ok {
		return claims
	}
	return nil
}

// GetCurrentUserID 从上下文获取当前用户ID
func GetCurrentUserID(c fiber.Ctx) int64 {
	if userID, ok := c.Locals("user_id").(int64); ok {
		return userID
	}
	return 0
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c fiber.Ctx) bool {
	return GetCurrentUser(c) != nil
}

// GetTokenFromContext 从上下文获取JWT令牌
func GetTokenFromContext(c fiber.Ctx) string {
	tokenString, _, validFormat := parseBearerToken(c.Get("Authorization"))
	if !validFormat {
		return ""
	}
	return tokenString
}
