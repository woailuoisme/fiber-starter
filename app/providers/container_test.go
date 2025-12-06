package providers

import (
	"testing"

	"fiber-starter/app/logger"
	"fiber-starter/app/services"
	"fiber-starter/config"
	"fiber-starter/database"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// initLogger 初始化日志系统（用于测试）
func initLogger() error {
	return logger.Init()
}

// TestNewContainer 测试创建容器
func TestNewContainer(t *testing.T) {
	container := NewContainer()
	if container == nil {
		t.Fatal("容器创建失败")
	}

	if container.Container == nil {
		t.Fatal("dig.Container 未初始化")
	}

	if len(container.providers) != 3 {
		t.Fatalf("期望 3 个 provider，实际得到 %d 个", len(container.providers))
	}
}

// TestRegisterProviders 测试注册所有 Provider
func TestRegisterProviders(t *testing.T) {
	// 初始化配置
	if err := config.Init(); err != nil {
		t.Skipf("跳过测试：配置初始化失败 - %v", err)
	}

	// 初始化日志（数据库连接需要）
	if err := initLogger(); err != nil {
		t.Skipf("跳过测试：日志初始化失败 - %v", err)
	}

	container := NewContainer()

	// 注册所有 Provider
	// 注意：这可能会失败如果数据库或 Redis 不可用，但这不影响 Provider 系统本身的正确性
	err := container.RegisterProviders()
	if err != nil {
		// 如果是数据库或 Redis 连接失败，跳过测试
		t.Skipf("跳过测试：注册 Provider 失败（可能是数据库或 Redis 不可用）- %v", err)
	}

	// 验证配置是否注册成功
	err = container.Invoke(func(cfg *config.Config) {
		if cfg == nil {
			t.Error("配置未成功注册")
		}
	})
	if err != nil {
		t.Errorf("获取配置失败: %v", err)
	}

	// 验证验证器是否注册成功
	err = container.Invoke(func(v *validator.Validate) {
		if v == nil {
			t.Error("验证器未成功注册")
		}
	})
	if err != nil {
		t.Errorf("获取验证器失败: %v", err)
	}

	// 验证数据库连接是否注册成功（如果数据库可用）
	err = container.Invoke(func(conn *database.Connection) {
		if conn == nil {
			t.Error("数据库连接未成功注册")
		}
	})
	if err != nil {
		t.Logf("数据库连接不可用（这是正常的如果数据库未运行）: %v", err)
	}

	// 验证 GORM 实例是否注册成功（如果数据库可用）
	err = container.Invoke(func(db *gorm.DB) {
		if db == nil {
			t.Error("GORM 实例未成功注册")
		}
	})
	if err != nil {
		t.Logf("GORM 实例不可用（这是正常的如果数据库未运行）: %v", err)
	}

	// 验证缓存服务是否注册成功（如果 Redis 可用）
	err = container.Invoke(func(cache services.CacheService) {
		if cache == nil {
			t.Error("缓存服务未成功注册")
		}
	})
	if err != nil {
		t.Logf("缓存服务不可用（这是正常的如果 Redis 未运行）: %v", err)
	}
}

// TestAppProvider 测试 AppProvider
func TestAppProvider(t *testing.T) {
	// 初始化配置
	if err := config.Init(); err != nil {
		t.Skipf("跳过测试：配置初始化失败 - %v", err)
	}

	container := NewContainer()
	appProvider := NewAppProvider(container.Container)

	if err := appProvider.Register(); err != nil {
		t.Fatalf("AppProvider 注册失败: %v", err)
	}

	// 验证配置
	err := container.Invoke(func(cfg *config.Config) {
		if cfg == nil {
			t.Error("配置未成功注册")
		}
		if cfg.App.Name == "" {
			t.Error("配置内容为空")
		}
	})
	if err != nil {
		t.Errorf("获取配置失败: %v", err)
	}

	// 验证验证器
	err = container.Invoke(func(v *validator.Validate) {
		if v == nil {
			t.Error("验证器未成功注册")
		}
	})
	if err != nil {
		t.Errorf("获取验证器失败: %v", err)
	}
}

// TestDatabaseProvider 测试 DatabaseProvider
func TestDatabaseProvider(t *testing.T) {
	// 初始化配置
	if err := config.Init(); err != nil {
		t.Skipf("跳过测试：配置初始化失败 - %v", err)
	}

	// 初始化日志（数据库连接需要）
	if err := initLogger(); err != nil {
		t.Skipf("跳过测试：日志初始化失败 - %v", err)
	}

	container := NewContainer()

	// 先注册 AppProvider（因为 DatabaseProvider 依赖配置）
	appProvider := NewAppProvider(container.Container)
	if err := appProvider.Register(); err != nil {
		t.Fatalf("AppProvider 注册失败: %v", err)
	}

	// 注册 DatabaseProvider
	dbProvider := NewDatabaseProvider(container.Container)
	if err := dbProvider.Register(); err != nil {
		t.Fatalf("DatabaseProvider 注册失败: %v", err)
	}

	// 验证数据库连接（如果数据库可用）
	err := container.Invoke(func(conn *database.Connection) {
		if conn == nil {
			t.Error("数据库连接未成功注册")
		}
		if conn.DB == nil {
			t.Error("GORM DB 实例为空")
		}
	})
	if err != nil {
		t.Skipf("跳过测试：数据库连接不可用 - %v", err)
	}

	// 验证 GORM 实例
	err = container.Invoke(func(db *gorm.DB) {
		if db == nil {
			t.Error("GORM 实例未成功注册")
		}
	})
	if err != nil {
		t.Skipf("跳过测试：GORM 实例不可用 - %v", err)
	}
}

// TestCacheProvider 测试 CacheProvider
func TestCacheProvider(t *testing.T) {
	// 初始化配置
	if err := config.Init(); err != nil {
		t.Skipf("跳过测试：配置初始化失败 - %v", err)
	}

	container := NewContainer()

	// 先注册 AppProvider（因为 CacheProvider 依赖配置）
	appProvider := NewAppProvider(container.Container)
	if err := appProvider.Register(); err != nil {
		t.Fatalf("AppProvider 注册失败: %v", err)
	}

	// 注册 CacheProvider
	cacheProvider := NewCacheProvider(container.Container)
	if err := cacheProvider.Register(); err != nil {
		t.Fatalf("CacheProvider 注册失败: %v", err)
	}

	// 验证 Redis 配置
	err := container.Invoke(func(redisCfg *config.RedisConfig) {
		if redisCfg == nil {
			t.Error("Redis 配置未成功注册")
		}
	})
	if err != nil {
		t.Errorf("获取 Redis 配置失败: %v", err)
	}

	// 验证缓存服务（即使 Redis 不可用，服务也应该被注册）
	err = container.Invoke(func(cache services.CacheService) {
		if cache == nil {
			t.Error("缓存服务未成功注册")
		}
	})
	if err != nil {
		t.Errorf("获取缓存服务失败: %v", err)
	}
}

// TestProviderInterface 测试 Provider 接口实现
func TestProviderInterface(t *testing.T) {
	container := NewContainer()

	// 验证所有 Provider 都实现了 Provider 接口
	providers := []Provider{
		NewAppProvider(container.Container),
		NewDatabaseProvider(container.Container),
		NewCacheProvider(container.Container),
	}

	for i, provider := range providers {
		if provider == nil {
			t.Errorf("Provider %d 为 nil", i)
		}
	}
}
