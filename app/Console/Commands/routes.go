package command

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	providers "fiber-starter/app/Providers"
	helpers "fiber-starter/app/Support"
	"fiber-starter/routes"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
)

var routesCmd = &cobra.Command{
	Use:   "routes",
	Short: "Display all registered routes",
	Long:  `List all registered route endpoints in the application, including HTTP methods and paths`,
	Run: func(_ *cobra.Command, _ []string) {
		showRoutes()
	},
}

func showRoutes() {
	app, runtime, err := setupRouteApp()
	if err != nil {
		return
	}
	defer func() {
		_ = runtime.Close()
		_ = helpers.Sync()
	}()

	allRoutes := app.GetRoutes()
	printRouteTable(allRoutes)
}

func setupRouteApp() (*fiber.App, *providers.Runtime, error) {
	runtime, err := buildRuntime()
	if err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to build runtime: %v\n", err)
		return nil, nil, err
	}

	app := fiber.New(fiber.Config{})
	if err := routes.SetupApplicationRoutes(
		app,
		runtime.Config,
		runtime.Cache,
		runtime.AuthController,
		runtime.UserController,
		runtime.HealthController,
	); err != nil {
		_, _ = color.New(color.FgRed).Printf("Failed to setup routes: %v\n", err)
		_ = runtime.Close()
		return nil, nil, err
	}

	return app, runtime, nil
}

type RouteInfo struct {
	Methods []string
	Path    string
	Handler string
}

func printRouteTable(allRoutes []fiber.Route) {
	routes := processRoutes(allRoutes)
	displayRoutes(routes, len(allRoutes))
}

func processRoutes(allRoutes []fiber.Route) []*RouteInfo {
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

	routes := make([]*RouteInfo, 0, len(routeMap))
	for _, route := range routeMap {
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

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	return routes
}

func displayRoutes(routes []*RouteInfo, totalEntries int) {
	fmt.Println()
	for _, route := range routes {
		printSingleRoute(route)
	}

	fmt.Println()
	_, _ = color.New(color.FgGreen).Printf("  Showing [%d] unique paths from [%d] route entries\n", len(routes), totalEntries)
	fmt.Println()
}

func printSingleRoute(route *RouteInfo) {
	methods := strings.Join(route.Methods, "|")

	methodColor := color.New(color.FgWhite)
	switch {
	case slices.Contains(route.Methods, "GET"):
		methodColor = color.New(color.FgGreen)
	case slices.Contains(route.Methods, "POST"):
		methodColor = color.New(color.FgYellow)
	case slices.Contains(route.Methods, "PUT") || slices.Contains(route.Methods, "PATCH"):
		methodColor = color.New(color.FgBlue)
	case slices.Contains(route.Methods, "DELETE"):
		methodColor = color.New(color.FgRed)
	}

	methodStr := fmt.Sprintf("%-12s", methods)
	pathStr := route.Path

	totalWidth := 80
	usedWidth := len(methods) + 2 + len(pathStr) + 1
	handler := route.Handler
	if handler != "" {
		usedWidth += len(handler) + 3
	}

	dots := 3
	if totalWidth > usedWidth {
		dots = totalWidth - usedWidth
	}

	_, _ = methodColor.Print(methodStr)
	fmt.Print("  ")
	_, _ = color.New(color.FgCyan).Print(pathStr)
	fmt.Print(" ")
	_, _ = color.New(color.FgWhite, color.Faint).Print(strings.Repeat(".", dots))
	if handler != "" {
		fmt.Print(" ")
		_, _ = color.New(color.FgWhite).Print(handler)
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(routesCmd)
}
