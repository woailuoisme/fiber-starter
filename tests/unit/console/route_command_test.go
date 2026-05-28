package tests

import (
	"bytes"
	"testing"

	routecmd "lfiber/internal/console/commands/route"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteCommandProcess_UsesRouteMetadata(t *testing.T) {
	routes := routecmd.Process([]fiber.Route{
		{Method: fiber.MethodHead, Path: "/api/v1/users", Name: ""},
		{Method: fiber.MethodGet, Path: "/api/v1/users", Name: "user:UserController.GetUsers"},
		{Method: fiber.MethodPost, Path: "/api/v1/auth/sign-in", Name: "auth:AuthController.SignIn"},
	})

	require.Len(t, routes, 2)
	byPath := routesByPath(routes)

	assert.Equal(t, []string{fiber.MethodGet, fiber.MethodHead}, byPath["/api/v1/users"].Methods)
	assert.Equal(t, "user", byPath["/api/v1/users"].Feature)
	assert.Equal(t, "UserController", byPath["/api/v1/users"].Controller)
	assert.Equal(t, "auth", byPath["/api/v1/auth/sign-in"].Feature)
	assert.Equal(t, "AuthController", byPath["/api/v1/auth/sign-in"].Controller)
}

func TestRouteCommandPrintTableWithGroup_Feature(t *testing.T) {
	var out bytes.Buffer
	routecmd.PrintTableWithGroup(&out, []fiber.Route{
		{Method: fiber.MethodGet, Path: "/api/v1/users", Name: "user:UserController.GetUsers"},
		{Method: fiber.MethodPost, Path: "/api/v1/auth/sign-in", Name: "auth:AuthController.SignIn"},
	}, routecmd.GroupFeature)

	rendered := out.String()
	assert.Contains(t, rendered, "  auth")
	assert.Contains(t, rendered, "  user")
	assert.Contains(t, rendered, "/api/v1/auth/sign-in")
	assert.Contains(t, rendered, "/api/v1/users")
}

func routesByPath(routes []*routecmd.Info) map[string]*routecmd.Info {
	result := make(map[string]*routecmd.Info, len(routes))
	for _, route := range routes {
		result[route.Path] = route
	}
	return result
}
