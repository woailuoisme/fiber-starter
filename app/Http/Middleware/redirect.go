package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
)

// SetupTimeoutRedirect 配置路径重定向中间件。
// 作用：在特定条件下将请求导向维护页、移动端页或旧路径新地址。
// 场景：站点维护、路径迁移、移动端首页重定向等。
// 使用方式：在基础中间件之后注册，内部根据请求路径和 User-Agent 决定跳转。
func SetupTimeoutRedirect(app *fiber.App) {
	app.Use(func(c fiber.Ctx) error {
		if redirectURL := resolveRedirectURL(c); redirectURL != "" {
			return c.Redirect().Status(fiber.StatusMovedPermanently).To(redirectURL)
		}
		return c.Next()
	})
}

func resolveRedirectURL(c fiber.Ctx) string {
	if isMaintenanceMode() && c.Path() != "/maintenance" {
		return "/maintenance"
	}

	if c.Path() == "/old-path" {
		return "/new-path"
	}

	userAgent := c.Get("User-Agent")
	if userAgent != "" && c.Path() == "/" && isMobile(userAgent) && !fiber.Query[bool](c, "desktop", false) {
		return "/mobile"
	}

	return ""
}

func isMobile(userAgent string) bool {
	mobileKeywords := []string{
		"Mobile", "Android", "iPhone", "iPad", "iPod", "BlackBerry",
		"Windows Phone", "webOS", "Opera Mini", "IEMobile",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, keyword := range mobileKeywords {
		if strings.Contains(userAgentLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func isMaintenanceMode() bool {
	return false
}
