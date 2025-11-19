package main

import (
	"embed"
	"html/template"
	"io/fs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/template/html"

	"fiber-starter/app/middleware"
)

//go:embed public/*
var publicFiles embed.FS

func main() {
	// 初始化HTML模板引擎
	engine := html.NewFileSystem(http.FS(publicFiles), ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// 设置中间件（包含favicon）
	middleware.SetupMiddleware(app)

	// 静态文件服务
	publicSubFS, _ := fs.Sub(publicFiles, "public")
	app.Static("/", publicSubFS)

	// 示例路由
	app.Get("/", func(c *fiber.Ctx) error {
		// 渲染HTML模板，包含favicon标签
		return c.Render("index", fiber.Map{
			"Title":       "Fiber Starter 应用",
			"FaviconTags": middleware.GetFaviconHTMLTags(),
		})
	})

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "服务运行正常",
		})
	})

	// 启动服务器
	app.Listen(":3000")
}
