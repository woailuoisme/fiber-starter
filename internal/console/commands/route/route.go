package route

import (
	"fmt"
	"sort"
	"strings"

	"lfiber/internal/bootstrap"
	"lfiber/internal/common/routing"
	"lfiber/internal/console/commands/commandutil"
	"lfiber/internal/console/ui"
	providers "lfiber/internal/providers"

	"github.com/gofiber/fiber/v3"
	"github.com/spf13/cobra"
)

type Info struct {
	Methods    []string
	Path       string
	Handler    string
	Feature    string
	Controller string
}

type GroupMode string

const (
	GroupNone       GroupMode = "none"
	GroupFeature    GroupMode = "feature"
	GroupController GroupMode = "controller"
)

func Commands() []*cobra.Command {
	return []*cobra.Command{listCommand()}
}

func listCommand() *cobra.Command {
	var groupBy string

	cmd := &cobra.Command{
		Use:     "route:list",
		Short:   "Display all registered routes",
		GroupID: "app",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			groupMode, err := parseGroupMode(groupBy)
			if err != nil {
				return err
			}

			app, runtime, err := setupRouteApp()
			if err != nil {
				return err
			}
			defer func() { _ = commandutil.CloseRuntime(runtime) }()
			PrintTableWithGroup(cmd.OutOrStdout(), app.GetRoutes(), groupMode)
			return nil
		},
	}
	cmd.Flags().StringVar(&groupBy, "group", string(GroupNone), "Group routes by none, feature, or controller")
	return cmd
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
	PrintTableWithGroup(out, allRoutes, GroupNone)
}

func PrintTableWithGroup(out interface{ Write([]byte) (int, error) }, allRoutes []fiber.Route, groupMode GroupMode) {
	routes := Process(allRoutes)
	_, _ = fmt.Fprintln(out)
	if groupMode == GroupNone {
		for _, route := range routes {
			printSingleRoute(out, route)
		}
	} else {
		for _, group := range groupedRoutes(routes, groupMode) {
			ui.Info(out, "  %s", group.Name)
			for _, route := range group.Routes {
				printSingleRoute(out, route)
			}
			_, _ = fmt.Fprintln(out)
		}
	}
	_, _ = fmt.Fprintln(out)
	ui.Success(out, "  Showing [%d] unique paths from [%d] route entries", len(routes), len(allRoutes))
	_, _ = fmt.Fprintln(out)
}

func Process(allRoutes []fiber.Route) []*Info {
	routeMap := make(map[string]*Info)
	for _, route := range allRoutes {
		key := route.Path
		meta := routing.MetadataFromRoute(route)
		if routeMap[key] == nil {
			routeMap[key] = &Info{
				Path:       route.Path,
				Handler:    route.Name,
				Feature:    meta.Feature,
				Controller: meta.Controller,
			}
		} else {
			mergeRouteMetadata(routeMap[key], route.Name, meta)
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

func mergeRouteMetadata(info *Info, handler string, meta routing.Metadata) {
	if info.Handler == "" && handler != "" {
		info.Handler = handler
	}
	if shouldReplaceGroup(info.Feature, meta.Feature) {
		info.Feature = meta.Feature
	}
	if shouldReplaceGroup(info.Controller, meta.Controller) {
		info.Controller = meta.Controller
	}
}

func shouldReplaceGroup(current, next string) bool {
	if strings.TrimSpace(next) == "" || next == routing.Unassigned {
		return false
	}
	return strings.TrimSpace(current) == "" || current == routing.Unassigned
}

type routeGroup struct {
	Name   string
	Routes []*Info
}

func parseGroupMode(value string) (GroupMode, error) {
	switch GroupMode(strings.ToLower(strings.TrimSpace(value))) {
	case "", GroupNone:
		return GroupNone, nil
	case GroupFeature:
		return GroupFeature, nil
	case GroupController:
		return GroupController, nil
	default:
		return GroupNone, fmt.Errorf("invalid route group %q, expected none, feature, or controller", value)
	}
}

func groupedRoutes(routes []*Info, mode GroupMode) []routeGroup {
	groupMap := make(map[string][]*Info)
	for _, route := range routes {
		name := routeGroupName(route, mode)
		groupMap[name] = append(groupMap[name], route)
	}

	names := make([]string, 0, len(groupMap))
	for name := range groupMap {
		names = append(names, name)
	}
	sort.Strings(names)

	groups := make([]routeGroup, 0, len(names))
	for _, name := range names {
		groups = append(groups, routeGroup{Name: name, Routes: groupMap[name]})
	}
	return groups
}

func routeGroupName(route *Info, mode GroupMode) string {
	if route == nil {
		return routing.Unassigned
	}
	var name string
	switch mode {
	case GroupFeature:
		name = route.Feature
	case GroupController:
		name = route.Controller
	}
	if strings.TrimSpace(name) == "" {
		return routing.Unassigned
	}
	return name
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
