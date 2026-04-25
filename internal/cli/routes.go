package command

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"fiber-starter/internal/app/providers"
	"fiber-starter/internal/config"
	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/transport/http/controllers"
	"fiber-starter/internal/transport/http/middleware"
	"fiber-starter/internal/transport/http/routers"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
)

// routesCmd Display all registered routes
var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Display all registered routes",
	Long:  `List all registered route endpoints in the application, including HTTP methods and paths`,
	Run: func(_ *cobra.Command, _ []string) {
		showRoutes()
	},
}

// showRoutes Display all routes
func showRoutes() {
	app, err := setupRouteApp()
	if err != nil {
		return
	}

	printRouteTable(app.GetRoutes())
}

// setupRouteApp Initialize app and setup routes
func setupRouteApp() (*fiber.App, error) {
	// Create dependency injection container
	container := providers.NewContainer()
	if err := container.RegisterProviders(); err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to register dependencies: %v\n", err)
		return nil, err
	}

	if err := container.Invoke(func(_ *config.Config) {}); err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to load config: %v\n", err)
		return nil, err
	}

	// Initialize logger (depends on loaded config, so must be after config loading)
	if err := helpers.Init(); err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to initialize logger: %v\n", err)
		return nil, err
	}

	// Create temporary Fiber app
	app := fiber.New(fiber.Config{})

	// Configure middleware and routes
	middleware.SetupMiddleware(app)
	middleware.SetupTimeoutRedirect(app)
	middleware.SetupAuthMiddleware(app)

	// Register routes
	err := container.Invoke(func(
		cfg *config.Config,
		cache helpers.CacheService,
		authController *controllers.AuthController,
		userController *controllers.UserController,
		healthController *controllers.HealthController,
	) {
		jwtProtected := middleware.JWTProtected(cfg, cache)
		routers.SetupRoutes(app, jwtProtected, authController, userController, healthController)
	})

	if err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to setup routes: %v\n", err)
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
	if slices.Contains(route.Methods, "GET") {
		methodColor = color.New(color.FgGreen)
	} else if slices.Contains(route.Methods, "POST") {
		methodColor = color.New(color.FgYellow)
	} else if slices.Contains(route.Methods, "PUT") || slices.Contains(route.Methods, "PATCH") {
		methodColor = color.New(color.FgBlue)
	} else if slices.Contains(route.Methods, "DELETE") {
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

// repeatStr 重复字符串
func repeatStr(s string, count int) string {
	return strings.Repeat(s, count)
}

func init() {
	rootCmd.AddCommand(routesCmd)
}
