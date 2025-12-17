package command

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"

	"fiber-starter/app/helpers"
	"fiber-starter/app/http/controllers"
	"fiber-starter/app/http/middleware"
	"fiber-starter/app/providers"
	"fiber-starter/app/routers"
	"fiber-starter/config"
)

// routesCmd 显示所有注册的路由
var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "显示所有注册的路由",
	Long:  `列出应用程序中所有注册的路由端点，包括 HTTP 方法和路径`,
	Run: func(cmd *cobra.Command, args []string) {
		showRoutes()
	},
}

// showRoutes 显示所有路由
func showRoutes() {
	// 初始化配置
	if err := config.Init(); err != nil {
		color.Red("初始化配置失败: %v", err)
		return
	}

	// 初始化日志
	if err := helpers.Init(); err != nil {
		color.Red("初始化日志失败: %v", err)
		return
	}

	// 创建依赖注入容器
	container := providers.NewContainer()
	if err := container.RegisterProviders(); err != nil {
		color.Red("注册依赖失败: %v", err)
		return
	}

	// 创建临时 Fiber 应用
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// 配置中间件和路由
	middleware.SetupMiddleware(app)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupErrorHandling(app)
	middleware.SetupAuthMiddleware(app)

	// 注册路由
	err := container.Invoke(func(authController *controllers.AuthController,
		userController *controllers.UserController,
		storageController *controllers.StorageController) {
		routers.SetupRoutes(app, authController, userController, storageController)
	})

	if err != nil {
		color.Red("设置路由失败: %v", err)
		return
	}

	// 获取所有路由
	allRoutes := app.GetRoutes()

	// 合并相同路径的不同方法
	type RouteInfo struct {
		Methods []string
		Path    string
		Handler string
	}

	routeMap := make(map[string]*RouteInfo)
	for _, route := range allRoutes {
		key := route.Path
		if routeMap[key] == nil {
			routeMap[key] = &RouteInfo{
				Path:    route.Path,
				Handler: route.Name,
			}
		}
		routeMap[key].Methods = append(routeMap[key].Methods, route.Method)
	}

	// 转换为切片并排序
	var routes []*RouteInfo
	for _, route := range routeMap {
		// 去重和排序方法
		methodSet := make(map[string]bool)
		for _, m := range route.Methods {
			methodSet[m] = true
		}
		route.Methods = []string{}
		for _, m := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
			if methodSet[m] {
				route.Methods = append(route.Methods, m)
			}
		}
		routes = append(routes, route)
	}

	// 按路径排序
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	// 打印路由（Laravel 风格）
	fmt.Println()
	for _, route := range routes {
		// 合并方法
		methods := ""
		if len(route.Methods) > 0 {
			methods = route.Methods[0]
			for i := 1; i < len(route.Methods); i++ {
				methods += "|" + route.Methods[i]
			}
		}

		// 方法颜色
		var methodColor *color.Color
		if contains(route.Methods, "GET") {
			methodColor = color.New(color.FgGreen)
		} else if contains(route.Methods, "POST") {
			methodColor = color.New(color.FgYellow)
		} else if contains(route.Methods, "PUT") || contains(route.Methods, "PATCH") {
			methodColor = color.New(color.FgBlue)
		} else if contains(route.Methods, "DELETE") {
			methodColor = color.New(color.FgRed)
		} else {
			methodColor = color.New(color.FgWhite)
		}

		// 处理器名称
		handler := route.Handler
		if handler == "" {
			handler = ""
		}

		// 格式化输出
		methodStr := fmt.Sprintf("%-12s", methods)
		pathStr := route.Path

		// 计算点的数量
		totalWidth := 120
		usedWidth := len(methods) + 2 + len(pathStr) + 1
		if handler != "" {
			usedWidth += len(handler) + 3
		}
		dots := totalWidth - usedWidth
		if dots < 3 {
			dots = 3
		}

		// 打印
		methodColor.Print(methodStr)
		fmt.Print("  ")
		color.New(color.FgCyan).Print(pathStr)
		fmt.Print(" ")
		color.New(color.FgWhite, color.Faint).Print(repeatStr(".", dots))
		if handler != "" {
			fmt.Print(" ")
			color.New(color.FgWhite).Print(handler)
		}
		fmt.Println()
	}

	// 打印统计
	fmt.Println()
	color.Green("  Showing [%d] routes\n", len(routes))
	fmt.Println()
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// repeatStr 重复字符串
func repeatStr(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func init() {
	rootCmd.AddCommand(routesCmd)
}
