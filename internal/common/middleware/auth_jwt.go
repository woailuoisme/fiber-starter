package middleware

import (
	"fmt"
	"reflect"
	"time"

	"lfiber/configs"
	exceptions "lfiber/internal/common/exceptions"
	auth "lfiber/internal/providers/auth"
	cacheContracts "lfiber/internal/providers/cache/contracts"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
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

// JWTOwner interface to avoid reflection for types that implement it
type JWTOwner interface {
	GetID() int64
	GetEmail() string
	GetName() string
}

// AuthUser defines the auth user schema in this middleware context
type AuthUser struct {
	ID    int64
	Email string
	Name  string
}

func (u *AuthUser) GetID() int64     { return u.ID }
func (u *AuthUser) GetEmail() string { return u.Email }
func (u *AuthUser) GetName() string  { return u.Name }

func setUserContext(c fiber.Ctx, claims *JWTClaims) {
	user := &AuthUser{
		ID:    claims.UserID,
		Email: claims.Email,
		Name:  claims.Name,
	}

	auth.SetUser(c, user)
	c.Locals("user_claims", claims)
}

// JWTAuth JWT authentication middleware
func JWTAuth(cfg *configs.Config) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.JWT.Secret)},
		Claims:     &JWTClaims{},
		Extractor:  extractors.FromAuthHeader(BearerSchema),
		ErrorHandler: func(c fiber.Ctx, err error) error {
			return exceptions.NewAuthenticationException(err.Error())
		},
		SuccessHandler: func(c fiber.Ctx) error {
			token := jwtware.FromContext(c)
			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				setUserContext(c, claims)
				return c.Next()
			}
			return exceptions.NewAuthenticationException("Invalid claims")
		},
	})
}

// OptionalJWTAuth 可选JWT认证中间件（不强制要求认证）
func OptionalJWTAuth(cfg *configs.Config) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(cfg.JWT.Secret)},
		Claims:     &JWTClaims{},
		Extractor:  extractors.FromAuthHeader(BearerSchema),
		ErrorHandler: func(c fiber.Ctx, _ error) error {
			// 对于可选认证，解析失败也继续执行后续逻辑
			return c.Next()
		},
		SuccessHandler: func(c fiber.Ctx) error {
			token := jwtware.FromContext(c)
			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				setUserContext(c, claims)
			}
			return c.Next()
		},
	})
}

// JWTProtected JWT middleware to protect routes (with blacklist check)
func JWTProtected(cfg *configs.Config, cacheStore cacheContracts.Store) fiber.Handler {
	jwtHandler := JWTAuth(cfg)

	return func(c fiber.Ctx) error {
		if cfg == nil {
			return exceptions.NewAPIException("Config not initialized", fiber.StatusInternalServerError)
		}

		// 1. 获取 Token
		tokenStr := GetTokenFromContext(c)
		if tokenStr == "" {
			return exceptions.NewAuthenticationException("Missing or invalid Authorization")
		}

		// 2. 检查黑名单 (退出登录后的 Token)
		if cacheStore != nil {
			blacklistKey := fmt.Sprintf("blacklist:%s", tokenStr)
			exists, err := cacheStore.Exists(blacklistKey)
			if err != nil {
				return exceptions.NewAPIException("Auth service unavailable", fiber.StatusServiceUnavailable)
			}
			if exists {
				return exceptions.NewAuthenticationException("Token has been invalidated")
			}
		}

		// 3. 执行官方 JWT 中间件校验
		return jwtHandler(c)
	}
}

func getInt64Field(obj any, name string) int64 {
	if obj == nil {
		return 0
	}
	if owner, ok := obj.(JWTOwner); ok {
		if name == "ID" {
			return owner.GetID()
		}
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		if f := val.FieldByName(name); f.IsValid() && f.Kind() == reflect.Int64 {
			return f.Int()
		}
	}
	return 0
}

func getStringField(obj any, name string) string {
	if obj == nil {
		return ""
	}
	if owner, ok := obj.(JWTOwner); ok {
		switch name {
		case "Email":
			return owner.GetEmail()
		case "Name":
			return owner.GetName()
		}
	}
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() == reflect.Struct {
		if f := val.FieldByName(name); f.IsValid() && f.Kind() == reflect.String {
			return f.String()
		}
	}
	return ""
}

// GenerateToken 生成JWT令牌
func GenerateToken(user any, cfg *configs.Config) (string, error) {
	claims := JWTClaims{
		UserID: getInt64Field(user, "ID"),
		Email:  getStringField(user, "Email"),
		Name:   getStringField(user, "Name"),
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
func GenerateRefreshToken(user any, cfg *configs.Config) (string, error) {
	claims := JWTClaims{
		UserID: getInt64Field(user, "ID"),
		Email:  getStringField(user, "Email"),
		Name:   getStringField(user, "Name"),
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
func ValidateToken(tokenString string, cfg *configs.Config) (*JWTClaims, error) {
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
func GetUserFromContext(c fiber.Ctx) any {
	return auth.User(c)
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
	return auth.Id(c)
}

// IsAuthenticated 检查用户是否已认证
func IsAuthenticated(c fiber.Ctx) bool {
	return auth.Check(c)
}

// GetTokenFromContext 从上下文获取JWT令牌
func GetTokenFromContext(c fiber.Ctx) string {
	token, _ := extractors.FromAuthHeader(BearerSchema).Extract(c)
	return token
}
