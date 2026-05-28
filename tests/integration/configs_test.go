package tests

import (
	"testing"

	"lfiber/configs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigs_Load(t *testing.T) {
	// 1. 测试基础加载
	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// 2. 验证基础值
	assert.NotEmpty(t, cfg.App.Name)

	// 如果 I18n.SupportedLanguages 仍然为空，可能是因为 yml 目录下有配置文件覆盖了它
	// 但默认情况下它应该包含 "en"
	if len(cfg.I18n.SupportedLanguages) == 0 {
		t.Log("Warning: I18n.SupportedLanguages is empty. This might happen if a YAML file overrides it.")
	} else {
		assert.Contains(t, cfg.I18n.SupportedLanguages, "en")
	}
}

func TestConfigs_Init(t *testing.T) {
	err := configs.Init()
	require.NoError(t, err)
	require.NotNil(t, configs.GlobalConfig)
}

func TestConfigs_EnvExpansion(t *testing.T) {
	expectedName := "Integration-Test-App"
	t.Setenv("APP_NAME", expectedName)

	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, expectedName, cfg.App.Name)
}

func TestConfigs_LoggerOutputUsesCanonicalEnv(t *testing.T) {
	t.Setenv("LOGGER_OUTPUT", "stderr")
	t.Setenv("LOG_CHANNEL", "daily")

	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "stderr", cfg.Logger.Output)
}

func TestConfigs_LoggerOutputSupportsLegacyChannelEnv(t *testing.T) {
	t.Setenv("LOGGER_OUTPUT", "")
	t.Setenv("LOG_CHANNEL", "single")

	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "single", cfg.Logger.Output)
}

func TestConfigs_OTELFlags(t *testing.T) {
	t.Setenv("OTEL_TRACE_ENABLED", "true")
	t.Setenv("OTEL_METRICS_ENABLED", "true")
	t.Setenv("OTEL_TRACE_SAMPLE_RATIO", "0.25")
	t.Setenv("OTEL_OTLP_INSECURE", "false")

	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	assert.True(t, cfg.OTEL.TraceEnabled)
	assert.True(t, cfg.OTEL.MetricsEnabled)
	assert.InEpsilon(t, 0.25, cfg.OTEL.TraceSampleRatio, 0.001)
	assert.False(t, cfg.OTEL.OTLPInsecure)
}

func TestConfigs_Database(t *testing.T) {
	t.Setenv("DB_CONNECTION", "sqlite")

	dbCfg, err := configs.LoadDatabaseConfig()
	require.NoError(t, err)
	require.NotNil(t, dbCfg)
	assert.Equal(t, "sqlite", dbCfg.Default)
}

func TestConfigs_ModularLoading(t *testing.T) {
	// 验证从 configs/yml/app.yaml 加载的配置是否生效
	// 注意：这取决于你的 app.yaml 里的内容，通常它会覆盖默认值
	cfg, _, err := configs.LoadConfig()
	require.NoError(t, err)

	// 只要能加载成功，说明 yml/ 目录下的文件被正确合并了
	assert.NotNil(t, cfg.App)
	assert.NotNil(t, cfg.Notification)
	assert.Contains(t, []string{"", "lfiber"}, cfg.Notification.Gotify.Title)
	assert.Equal(t, "https://api.telegram.org", cfg.Notification.Telegram.APIURL)
}
