package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
)

// APIKeyAuth 返回用于内部调用的 KeyAuth 中间件。
// 作用：校验 API Key 或 Bearer token，保护内部或机器到机器接口。
// 场景：内部管理接口、任务回调、服务间调用、临时受控开放接口。
// 使用方式：在路由组上挂载，并通过 X-API-Key 或 Authorization: Bearer 传入密钥。
func APIKeyAuth(apiKey string) fiber.Handler {
	secret := strings.TrimSpace(apiKey)
	return keyauth.New(keyauth.Config{
		Next: func(c fiber.Ctx) bool {
			return secret == ""
		},
		Extractor: extractors.Chain(
			extractors.FromHeader("X-API-Key"),
			extractors.FromAuthHeader("Bearer"),
		),
		Validator: func(_ fiber.Ctx, key string) (bool, error) {
			if secret == "" {
				return false, keyauth.ErrMissingOrMalformedAPIKey
			}
			if subtle.ConstantTimeCompare([]byte(key), []byte(secret)) == 1 {
				return true, nil
			}
			return false, keyauth.ErrMissingOrMalformedAPIKey
		},
	})
}
