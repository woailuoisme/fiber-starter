package providers_test

import (
	"testing"

	"fiber-starter/configs"
	logging "fiber-starter/internal/providers/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingProvider_Register(t *testing.T) {
	service, err := logging.Register(configs.LoggerConfig{
		Level:  "info",
		Output: "stdout",
	})
	require.NoError(t, err)

	require.NotNil(t, service)
	require.NotNil(t, service.Default())

	assert.Equal(t, service.GetZapLogger(), service.Channel("default").GetZapLogger())
	assert.Equal(t, service.GetZapLogger(), service.Channel("unknown-channel").GetZapLogger())
}
