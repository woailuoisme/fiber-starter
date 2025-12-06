package tests

import (
	"fiber-starter/config"
	"fiber-starter/database"
	"testing"
)

// TestDatabaseConnection 测试数据库连接
func TestDatabaseConnection(t *testing.T) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("跳过测试：无法加载配置 - %v", err)
		return
	}

	// 创建数据库连接
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("跳过测试：无法连接数据库 - %v", err)
		return
	}
	defer conn.Close()

	// 测试连接是否成功
	if conn.DB == nil {
		t.Fatal("数据库连接为空")
	}

	t.Log("数据库连接成功")
}

// TestDatabaseHealthCheck 测试数据库健康检查
func TestDatabaseHealthCheck(t *testing.T) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("跳过测试：无法加载配置 - %v", err)
		return
	}

	// 创建数据库连接
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("跳过测试：无法连接数据库 - %v", err)
		return
	}
	defer conn.Close()

	// 执行健康检查
	err = conn.HealthCheck()
	if err != nil {
		t.Fatalf("数据库健康检查失败: %v", err)
	}

	t.Log("数据库健康检查通过")
}

// TestDatabaseConnectionStats 测试获取数据库连接池统计信息
func TestDatabaseConnectionStats(t *testing.T) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("跳过测试：无法加载配置 - %v", err)
		return
	}

	// 创建数据库连接
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("跳过测试：无法连接数据库 - %v", err)
		return
	}
	defer conn.Close()

	// 获取连接池统计信息
	stats, err := conn.GetStats()
	if err != nil {
		t.Fatalf("获取连接池统计信息失败: %v", err)
	}

	// 验证统计信息包含必要的字段
	requiredFields := []string{
		"max_open_connections",
		"open_connections",
		"in_use",
		"idle",
		"wait_count",
		"wait_duration",
	}

	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("统计信息缺少字段: %s", field)
		}
	}

	t.Logf("连接池统计信息: %+v", stats)
}

// TestDatabaseConnectionPool 测试数据库连接池配置
func TestDatabaseConnectionPool(t *testing.T) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("跳过测试：无法加载配置 - %v", err)
		return
	}

	// 创建数据库连接
	conn, err := database.NewConnection(cfg)
	if err != nil {
		t.Skipf("跳过测试：无法连接数据库 - %v", err)
		return
	}
	defer conn.Close()

	// 获取底层 sql.DB
	sqlDB, err := conn.DB.DB()
	if err != nil {
		t.Fatalf("获取底层sql.DB失败: %v", err)
	}

	// 验证连接池配置
	stats := sqlDB.Stats()

	if stats.MaxOpenConnections != cfg.Database.Pool.MaxOpenConns {
		t.Errorf("最大打开连接数配置不匹配: 期望 %d, 实际 %d",
			cfg.Database.Pool.MaxOpenConns, stats.MaxOpenConnections)
	}

	t.Logf("连接池配置验证通过 - MaxOpenConns: %d, MaxIdleConns: %d",
		cfg.Database.Pool.MaxOpenConns, cfg.Database.Pool.MaxIdleConns)
}

// TestGlobalHealthCheck 测试全局健康检查函数
func TestGlobalHealthCheck(t *testing.T) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("跳过测试：无法加载配置 - %v", err)
		return
	}

	// 创建数据库连接并设置全局实例
	_, err = database.NewConnection(cfg)
	if err != nil {
		t.Skipf("跳过测试：无法连接数据库 - %v", err)
		return
	}

	// 测试全局健康检查函数
	err = database.HealthCheck()
	if err != nil {
		t.Fatalf("全局健康检查失败: %v", err)
	}

	t.Log("全局健康检查通过")
}

// TestGetConnectionStats 测试全局获取连接统计信息函数
func TestGetConnectionStats(t *testing.T) {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Skipf("跳过测试：无法加载配置 - %v", err)
		return
	}

	// 创建数据库连接并设置全局实例
	_, err = database.NewConnection(cfg)
	if err != nil {
		t.Skipf("跳过测试：无法连接数据库 - %v", err)
		return
	}

	// 测试全局获取连接统计信息函数
	stats, err := database.GetConnectionStats()
	if err != nil {
		t.Fatalf("获取连接统计信息失败: %v", err)
	}

	if len(stats) == 0 {
		t.Fatal("连接统计信息为空")
	}

	t.Logf("全局连接统计信息: %+v", stats)
}
