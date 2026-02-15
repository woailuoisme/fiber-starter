package i18n

import (
	"strings"
	"time"

	"fiber-starter/app/helpers"
	"fiber-starter/config"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Middleware 语言检测中间件
// 按优先级检测用户语言：Query > Cookie > Accept-Language Header > Default
func Middleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 检测用户语言
		lang := detectLanguage(c)

		// 创建翻译器并设置到上下文
		translator := NewTranslator(lang)
		SetToContext(c, translator)

		// 如果语言来自 query 参数，设置 Cookie
		if c.Query("lang") != "" {
			setLanguageCookie(c, lang)
		}

		return c.Next()
	}
}

// detectLanguage 检测用户语言偏好
// 优先级：Query > Cookie > Accept-Language Header > Default
func detectLanguage(c fiber.Ctx) string {
	// 1. 检查 query 参数
	if lang := c.Query("lang"); lang != "" {
		if IsSupported(lang) {
			helpers.Debug("从 query 参数检测到语言", zap.String("lang", lang))
			return lang
		}
		helpers.Warn("query 参数中的语言不支持", zap.String("lang", lang))
	}

	// 2. 检查 Cookie
	cfg := config.GlobalConfig.I18n
	if lang := c.Cookies(cfg.CookieName); lang != "" {
		if IsSupported(lang) {
			helpers.Debug("从 Cookie 检测到语言", zap.String("lang", lang))
			return lang
		}
		helpers.Warn("Cookie 中的语言不支持", zap.String("lang", lang))
	}

	// 3. 检查 Accept-Language 请求头
	if acceptLang := c.Get("Accept-Language"); acceptLang != "" {
		langs := parseAcceptLanguage(acceptLang)
		for _, lang := range langs {
			if IsSupported(lang) {
				helpers.Debug("从 Accept-Language 检测到语言", zap.String("lang", lang))
				return lang
			}
		}
		helpers.Debug("Accept-Language 中没有支持的语言", zap.String("accept-language", acceptLang))
	}

	// 4. 使用默认语言
	helpers.Debug("使用默认语言", zap.String("lang", DefaultLanguage))
	return DefaultLanguage
}

// parseAcceptLanguage 解析 Accept-Language 请求头
// 返回按优先级排序的语言列表
func parseAcceptLanguage(header string) []string {
	if header == "" {
		return []string{}
	}

	var languages []string

	// 分割多个语言
	parts := strings.Split(header, ",")
	for _, part := range parts {
		// 移除空格
		part = strings.TrimSpace(part)

		// 分割语言和权重 (例如: zh-CN;q=0.9)
		langParts := strings.Split(part, ";")
		if len(langParts) > 0 {
			lang := strings.TrimSpace(langParts[0])

			// 标准化语言代码
			lang = normalizeLanguageCode(lang)

			if lang != "" {
				languages = append(languages, lang)
			}
		}
	}

	return languages
}

// normalizeLanguageCode 标准化语言代码
// 例如：zh-cn -> zh-CN, en-us -> en
func normalizeLanguageCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}

	// 转换为小写
	code = strings.ToLower(code)

	// 处理常见的语言代码
	switch {
	case strings.HasPrefix(code, "zh-cn") || code == "zh":
		return "zh-CN"
	case strings.HasPrefix(code, "zh-tw") || strings.HasPrefix(code, "zh-hk"):
		return "zh-TW"
	case strings.HasPrefix(code, "en"):
		return "en"
	case strings.HasPrefix(code, "ja"):
		return "ja"
	case strings.HasPrefix(code, "ko"):
		return "ko"
	default:
		// 对于其他语言，保持原样但首字母大写
		parts := strings.Split(code, "-")
		if len(parts) == 2 {
			return parts[0] + "-" + strings.ToUpper(parts[1])
		}
		return code
	}
}

// setLanguageCookie 设置语言 Cookie
func setLanguageCookie(c fiber.Ctx, lang string) {
	cfg := config.GlobalConfig.I18n

	c.Cookie(&fiber.Cookie{
		Name:     cfg.CookieName,
		Value:    lang,
		MaxAge:   cfg.CookieMaxAge,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Expires:  time.Now().Add(time.Duration(cfg.CookieMaxAge) * time.Second),
	})

	helpers.Debug("设置语言 Cookie", zap.String("lang", lang))
}

// GetLanguageCookie 获取语言 Cookie
func GetLanguageCookie(c fiber.Ctx) string {
	cfg := config.GlobalConfig.I18n
	return c.Cookies(cfg.CookieName)
}

// ClearLanguageCookie 清除语言 Cookie
func ClearLanguageCookie(c fiber.Ctx) {
	cfg := config.GlobalConfig.I18n

	c.Cookie(&fiber.Cookie{
		Name:     cfg.CookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		Expires:  time.Now().Add(-time.Hour),
	})

	helpers.Debug("清除语言 Cookie")
}

// SetLanguage 设置用户语言（通过 Cookie）
func SetLanguage(c fiber.Ctx, lang string) error {
	if !IsSupported(lang) {
		return fiber.NewError(fiber.StatusBadRequest, "不支持的语言: "+lang)
	}

	setLanguageCookie(c, lang)

	// 更新当前请求的翻译器
	translator := NewTranslator(lang)
	SetToContext(c, translator)

	return nil
}

// GetCurrentLanguage 获取当前请求的语言
func GetCurrentLanguage(c fiber.Ctx) string {
	translator := GetFromContext(c)
	if translator != nil {
		return translator.GetLanguage()
	}
	return DefaultLanguage
}
