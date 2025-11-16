package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"fiber-starter/app/controllers"
	"fiber-starter/app/helpers"
	"fiber-starter/app/routers"
	"fiber-starter/app/services"
	"fiber-starter/config"

	_ "fiber-starter/docs" // swagger docs
)

// @title Fiber Starter API
// @version 1.0
// @description 这是一个基于 Fiber 框架的 Go 项目启动模板 API 文档
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// 初始化配置
	err := config.Init()
	if err != nil {
		return
	}

	// 创建 Fiber 应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(helpers.ErrorResponse(err.Error(), nil))
		},
	})

	// 中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// 初始化数据库连接
	// 构建数据库连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=%s",
		config.GetString("database.username"),
		config.GetString("database.password"),
		config.GetString("database.host"),
		config.GetString("database.port"),
		config.GetString("database.database"),
		config.GetString("database.charset"),
		config.GetString("database.timezone"))
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 初始化验证器
	validate := validator.New()

	// 初始化缓存服务
	cacheService := services.NewCacheService(config.GlobalConfig)

	// 初始化认证服务
	authService := services.NewAuthService(db, config.GlobalConfig, cacheService)

	// 初始化用户服务
	userService := services.NewUserService(db)

	// 初始化控制器
	authController := controllers.NewAuthController(authService, validate)
	userController := controllers.NewUserController(userService, validate)

	// 配置路由
	routers.SetupRoutes(app, authController, userController)

	// 启动服务器
	port := ":" + config.GetString("app.port")
	log.Printf("服务器启动在端口 %s", port)
	log.Fatal(app.Listen(port))
}
