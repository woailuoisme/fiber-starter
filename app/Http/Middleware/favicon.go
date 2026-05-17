package middleware

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/favicon"
)

// SetupFavicon 挂载站点图标中间件和静态路由。
// 作用：为浏览器自动请求的图标提供响应，减少 404 噪音。
// 场景：本地开发、管理后台、直接打开 API 域名。
// 使用方式：存在 public/favicon.ico 时自动启用，同时暴露 /favicon.svg。
func SetupFavicon(app *fiber.App) {
	if _, err := os.Stat("./public/favicon.ico"); err == nil {
		app.Use(favicon.New(favicon.Config{
			File: "./public/favicon.ico",
			URL:  "/favicon.ico",
		}))
	}

	app.Get("/favicon.svg", func(c fiber.Ctx) error {
		return c.SendFile("./public/favicon.svg")
	})
}
