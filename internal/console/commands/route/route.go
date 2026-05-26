package route

import (
	"fmt"
	"sort"
	"strings"

	"lfiber/internal/bootstrap"
	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/ui"
	providers "lfiber/internal/providers"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
)

type Info struct {
	Methods []string
	Path    string
	Handler string
}

func Commands() []*cobra.Command {
	return []*cobra.Command{listCommand()}
}

func listCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "route:list",
		Short:   "Display all registered routes",
		GroupID: "app",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			app, runtime, err := setupRouteApp()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(runtime) }()
			PrintTable(cmd.OutOrStdout(), app.GetRoutes())
			return nil
		},
	}
}

func setupRouteApp() (*fiber.App, *providers.Runtime, error) {
	runtime, err := commandutil.BuildRuntime()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build runtime: %w", err)
	}

	app := fiber.New(fiber.Config{})
	if err := bootstrap.SetupApplicationRoutes(app); err != nil {
		_ = commandutil.CloseRuntime(runtime)
		return nil, nil, fmt.Errorf("failed to setup routes: %w", err)
	}
	return app, runtime, nil
}

func PrintTable(out interface{ Write([]byte) (int, error) }, allRoutes []fiber.Route) {
	routes := Process(allRoutes)
	_, _ = fmt.Fprintln(out)
	for _, route := range routes {
		printSingleRoute(out, route)
	}
	_, _ = fmt.Fprintln(out)
	ui.Success(out, "  Showing [%d] unique paths from [%d] route entries", len(routes), len(allRoutes))
	_, _ = fmt.Fprintln(out)
}

func Process(allRoutes []fiber.Route) []*Info {
	routeMap := make(map[string]*Info)
	for _, route := range allRoutes {
		key := route.Path
		if routeMap[key] == nil {
			routeMap[key] = &Info{Path: route.Path, Handler: route.Name}
		}
		routeMap[key].Methods = append(routeMap[key].Methods, route.Method)
	}

	routes := make([]*Info, 0, len(routeMap))
	for _, route := range routeMap {
		methodSet := make(map[string]bool)
		for _, method := range route.Methods {
			methodSet[method] = true
		}

		route.Methods = []string{}
		for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} {
			if methodSet[method] {
				route.Methods = append(route.Methods, method)
			}
		}
		routes = append(routes, route)
	}

	sort.Slice(routes, func(i, j int) bool { return routes[i].Path < routes[j].Path })
	return routes
}

func printSingleRoute(out interface{ Write([]byte) (int, error) }, route *Info) {
	methods := strings.Join(route.Methods, "|")
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

	_, _ = fmt.Fprint(out, ui.Method(out, route.Methods))
	_, _ = fmt.Fprint(out, "  ")
	_, _ = fmt.Fprint(out, ui.Highlight(out, pathStr))
	_, _ = fmt.Fprint(out, " ")
	_, _ = fmt.Fprint(out, ui.Faint(out, strings.Repeat(".", dots)))
	if handler != "" {
		_, _ = fmt.Fprint(out, " ")
		_, _ = fmt.Fprint(out, handler)
	}
	_, _ = fmt.Fprintln(out)
}
