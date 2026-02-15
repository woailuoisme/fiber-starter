package tests

import (
	"bytes"
	"testing"
	"time"

	"fiber-starter/app/services"
	"fiber-starter/config"
)

// TestStorageService 测试存储服务功能
func TestStorageService(t *testing.T) {
	// 测试内存存储
	t.Run("MemoryStorage", func(t *testing.T) {
		testStorageDriver(t, "memory")
	})

	// 测试BBolt存储
	t.Run("BBoltStorage", func(t *testing.T) {
		testStorageDriver(t, "bbolt")
	})

	// 测试Redis存储
	t.Run("RedisStorage", func(t *testing.T) {
		testStorageDriver(t, "redis")
	})
}

// testStorageDriver 测试指定驱动的存储功能
func testStorageDriver(t *testing.T, driver string) {
	// 创建存储服务
	storageService, err := createTestStorageService(driver)
	if err != nil {
		t.Fatalf("创建存储服务失败: %v", err)
	}

	testKey := "test_key_" + driver
	testValue := []byte("test_value_" + driver)

	performStorageOperations(t, storageService, testKey, testValue)
}

// createTestStorageService 创建测试用的存储服务
func createTestStorageService(driver string) (*services.StorageService, error) {
	storageCfg := &config.StorageConfig{
		Driver:   driver,
		Database: "./test_storage.db",
		Reset:    true,
	}
	redisCfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "", // 无密码
		DB:       0,
	}

	return services.NewStorageService(storageCfg, redisCfg)
}

// performStorageOperations 执行存储操作测试
func performStorageOperations(t *testing.T, storageService *services.StorageService, testKey string, testValue []byte) {
	// 设置值
	err := storageService.Set(testKey, testValue, time.Minute)
	if err != nil {
		t.Errorf("设置存储值失败: %v", err)
	}

	// 获取值
	value, err := storageService.Get(testKey)
	if err != nil {
		t.Errorf("获取存储值失败: %v", err)
	}

	if !bytes.Equal(value, testValue) {
		t.Errorf("值不匹配，期望 %s，实际 %s", string(testValue), string(value))
	}

	// 检查键是否存在
	exists, err := storageService.Exists(testKey)
	if err != nil {
		t.Errorf("检查键存在性失败: %v", err)
	}

	if !exists {
		t.Errorf("键应该存在，但返回不存在")
	}

	// 删除键
	err = storageService.Delete(testKey)
	if err != nil {
		t.Errorf("删除键失败: %v", err)
	}

	// 再次检查键是否存在
	exists, err = storageService.Exists(testKey)
	if err != nil {
		t.Errorf("检查键存在性失败: %v", err)
	}

	if exists {
		t.Errorf("键应该不存在，但返回存在")
	}

	// 测试获取不存在的键
	_, err = storageService.Get("non_existent_key")
	if err == nil {
		t.Errorf("获取不存在的键应该返回错误")
	}
}

// TestStorageServiceExpire 测试存储过期功能
func TestStorageServiceExpire(t *testing.T) {
	// 创建临时配置
	storageCfg := &config.StorageConfig{
		Driver:   "memory",
		Database: "./test_storage.db",
		Reset:    true,
	}
	redisCfg := &config.RedisConfig{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		DB:       0,
	}

	// 创建存储服务
	storageService, err := services.NewStorageService(storageCfg, redisCfg)
	if err != nil {
		t.Fatalf("创建存储服务失败: %v", err)
	}

	testKey := "expire_test_key"
	testValue := []byte("expire_test_value")

	// 设置一个短时间的值（1秒）
	err = storageService.Set(testKey, testValue, time.Second)
	if err != nil {
		t.Fatalf("设置存储值失败: %v", err)
	}

	// 立即获取应该存在
	_, err = storageService.Get(testKey)
	if err != nil {
		t.Errorf("立即获取存储值失败: %v", err)
	}

	// 等待2秒后应该过期
	time.Sleep(2 * time.Second)

	_, err = storageService.Get(testKey)
	if err == nil {
		t.Errorf("过期的键应该返回错误")
	}

	t.Log("存储过期功能测试通过")
}
