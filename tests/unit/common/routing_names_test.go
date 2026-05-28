package common_test

import (
	"testing"

	"lfiber/internal/common/routing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestRoutingMetadataFromFunctionName_UsesFeaturePackageAndControllerReceiver(t *testing.T) {
	meta := routing.MetadataFromFunctionName("lfiber/internal/features/user.(*UserController).GetUsers-fm")

	assert.Equal(t, "user", meta.Feature)
	assert.Equal(t, "UserController", meta.Controller)
}

func TestRoutingMetadataFromFunctionName_IgnoresAnonymousFunctionsAsControllers(t *testing.T) {
	meta := routing.MetadataFromFunctionName("lfiber/internal/features/monitoring.RegisterRoutes.func1")

	assert.Equal(t, "monitoring", meta.Feature)
	assert.Empty(t, meta.Controller)
}

func TestRoutingMetadataFromName_SplitsFeatureAndController(t *testing.T) {
	meta := routing.MetadataFromName("user:UserController.GetUsers")

	assert.Equal(t, "user", meta.Feature)
	assert.Equal(t, "UserController", meta.Controller)
}

func TestRoutingMetadataFromRoute_FallsBackToPathFeature(t *testing.T) {
	meta := routing.MetadataFromRoute(fiber.Route{Path: "/api/v1/users"})

	assert.Equal(t, "user", meta.Feature)
	assert.Equal(t, routing.Unassigned, meta.Controller)
}
