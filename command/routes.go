package command

import (
	"fmt"
	"sort"
	"strings"

	"fiber-starter/app/helpers"
	"fiber-starter/app/http/controllers"
	"fiber-starter/app/http/middleware"
	"fiber-starter/app/providers"
	"fiber-starter/app/routers"
	"fiber-starter/config"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
)

// routesCmd 显示所有注册的路由
var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "显示所有注册的路由",
	Long:  `列出应用程序中所有注册的路由端点，包括 HTTP 方法和路径`,
	Run: func(_ *cobra.Command, _ []string) {
		showRoutes()
	},
}

// showRoutes 显示所有路由
func showRoutes() {
	app, err := setupRouteApp()
	if err != nil {
		return
	}

	printRouteTable(app.GetRoutes())
}

// setupRouteApp 初始化应用并设置路由
func setupRouteApp() (*fiber.App, error) {
	// 初始化配置
	if err := config.Init(); err != nil {
		_, _ = color.New(color.FgRed).Printf("初始化配置失败: %v\n", err)
		return nil, err
	}

	// 初始化日志
	if err := helpers.Init(); err != nil {
		_, _ = color.New(color.FgRed).Printf("初始化日志失败: %v\n", err)
		return nil, err
	}

	// 创建依赖注入容器
	container := providers.NewContainer()
	if err := container.RegisterProviders(); err != nil {
		_, _ = color.New(color.FgRed).Printf("注册依赖失败: %v\n", err)
		return nil, err
	}

	// 创建临时 Fiber 应用
	app := fiber.New(fiber.Config{})

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
		_, _ = color.New(color.FgRed).Printf("设置路由失败: %v\n", err)
		return nil, err
	}

	return app, nil
}

// RouteInfo 路由信息结构体
type RouteInfo struct {
	Methods []string
	Path    string
	Handler string
}

// printRouteTable 打印路由表
func printRouteTable(allRoutes []fiber.Route) {
	routes := processRoutes(allRoutes)
	displayRoutes(routes)
}

// processRoutes 处理和排序路由
func processRoutes(allRoutes []fiber.Route) []*RouteInfo {
	// 合并相同路径的不同方法
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

	// 转换为切片并处理方法
	routes := make([]*RouteInfo, 0, len(routeMap))
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

	return routes
}

// displayRoutes 显示路由列表
func displayRoutes(routes []*RouteInfo) {
	fmt.Println()
	for _, route := range routes {
		printSingleRoute(route)
	}

	// 打印统计
	fmt.Println()
	_, _ = color.New(color.FgGreen).Printf("  Showing [%d] routes\n", len(routes))
	fmt.Println()
}

// printSingleRoute 打印单个路由信息
func printSingleRoute(route *RouteInfo) {
	methods := strings.Join(route.Methods, "|")

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

	// 格式化输出
	methodStr := fmt.Sprintf("%-12s", methods)
	pathStr := route.Path

	// 计算点的数量
	totalWidth := 80 // 稍微减小宽度以适应大多数终端
	usedWidth := len(methods) + 2 + len(pathStr) + 1
	handler := route.Handler
	if handler != "" {
		usedWidth += len(handler) + 3
	}

	dots := 3
	if totalWidth > usedWidth {
		dots = totalWidth - usedWidth
	}

	// 打印
	_, _ = methodColor.Print(methodStr)
	fmt.Print("  ")
	_, _ = color.New(color.FgCyan).Print(pathStr)
	fmt.Print(" ")
	_, _ = color.New(color.FgWhite, color.Faint).Print(repeatStr(".", dots))
	if handler != "" {
		fmt.Print(" ")
		_, _ = color.New(color.FgWhite).Print(handler)
	}
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
