package main

import (
	"fmt"
	"strings"
	"time"

	"fiber-starter/app/services"
	"fiber-starter/config"
)

// MinIO和S3存储演示
func main() {
	fmt.Println("=== MinIO 和 S3 存储演示 ===\n")

	// 演示MinIO存储
	demonstrateMinIO()

	fmt.Println("\n" + strings.Repeat("-", 50) + "\n")

	// 演示S3存储
	demonstrateS3()
}

// demonstrateMinIO 演示MinIO存储功能
func demonstrateMinIO() {
	fmt.Println("1. MinIO 存储演示")
	fmt.Println("================")

	// 配置MinIO存储
	minioConfig := &config.MinIOStorageConfig{
		Endpoint:        "localhost:9000",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		UseSSL:          false,
		Bucket:          "fiber-starter-test",
		Region:          "us-east-1",
	}

	storageConfig := &config.StorageConfig{
		Driver: "minio",
		MinIO:  minioConfig,
	}

	// 创建MinIO存储服务
	storageService, err := services.NewStorageService(storageConfig, nil)
	if err != nil {
		fmt.Printf("创建MinIO存储服务失败: %v\n", err)
		fmt.Println("提示: 请确保MinIO服务正在运行在 localhost:9000")
		return
	}
	defer storageService.Close()

	// 测试基本存储操作
	testBasicStorageOperations(storageService, "MinIO")

	// 测试字符串操作
	testStringOperations(storageService, "MinIO")
}

// demonstrateS3 演示S3存储功能
func demonstrateS3() {
	fmt.Println("2. S3 存储演示")
	fmt.Println("==============")

	// 配置S3存储
	s3Config := &config.S3StorageConfig{
		AccessKeyID:     "your-access-key-id",
		SecretAccessKey: "your-secret-access-key",
		Region:          "us-east-1",
		Bucket:          "fiber-starter-test",
		Endpoint:        "", // 留空使用AWS S3，或设置自定义端点
	}

	storageConfig := &config.StorageConfig{
		Driver: "s3",
		S3:     s3Config,
	}

	// 创建S3存储服务
	storageService, err := services.NewStorageService(storageConfig, nil)
	if err != nil {
		fmt.Printf("创建S3存储服务失败: %v\n", err)
		fmt.Println("提示: 请配置正确的AWS凭证或使用本地S3兼容服务")
		return
	}
	defer storageService.Close()

	// 测试基本存储操作
	testBasicStorageOperations(storageService, "S3")

	// 测试字符串操作
	testStringOperations(storageService, "S3")
}

// testBasicStorageOperations 测试基本存储操作
func testBasicStorageOperations(storageService *services.StorageService, provider string) {
	fmt.Printf("\n%s 基本存储操作测试:\n", provider)

	// 测试设置和获取值
	testKey := "test-key"
	testValue := []byte("Hello, " + provider + " Storage!")

	fmt.Printf("设置键值对: %s = %s\n", testKey, string(testValue))
	err := storageService.Set(testKey, testValue, time.Hour)
	if err != nil {
		fmt.Printf("设置失败: %v\n", err)
		return
	}

	// 获取值
	retrievedValue, err := storageService.Get(testKey)
	if err != nil {
		fmt.Printf("获取失败: %v\n", err)
		return
	}

	fmt.Printf("获取值: %s\n", string(retrievedValue))

	// 检查键是否存在
	exists, err := storageService.Exists(testKey)
	if err != nil {
		fmt.Printf("检查存在性失败: %v\n", err)
		return
	}
	fmt.Printf("键 '%s' 存在: %t\n", testKey, exists)

	// 设置过期时间
	fmt.Printf("设置键 '%s' 的过期时间为30秒\n", testKey)
	err = storageService.SetExpire(testKey, 30*time.Second)
	if err != nil {
		fmt.Printf("设置过期时间失败: %v\n", err)
	}

	// 删除键
	fmt.Printf("删除键 '%s'\n", testKey)
	err = storageService.Delete(testKey)
	if err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}

	// 再次检查键是否存在
	exists, err = storageService.Exists(testKey)
	if err != nil {
		fmt.Printf("检查存在性失败: %v\n", err)
		return
	}
	fmt.Printf("删除后键 '%s' 存在: %t\n", testKey, exists)
}

// testStringOperations 测试字符串操作
func testStringOperations(storageService *services.StorageService, provider string) {
	fmt.Printf("\n%s 字符串操作测试:\n", provider)

	// 测试字符串设置和获取
	stringKey := "string-test"
	stringValue := fmt.Sprintf("这是一个测试字符串 - %s", provider)

	fmt.Printf("设置字符串: %s = %s\n", stringKey, stringValue)
	err := storageService.SetString(stringKey, stringValue, time.Hour)
	if err != nil {
		fmt.Printf("设置字符串失败: %v\n", err)
		return
	}

	// 获取字符串
	retrievedString, err := storageService.GetString(stringKey)
	if err != nil {
		fmt.Printf("获取字符串失败: %v\n", err)
		return
	}

	fmt.Printf("获取字符串: %s\n", retrievedString)

	// 使用默认TTL设置字符串
	defaultTTLKey := "default-ttl-test"
	defaultTTLValue := fmt.Sprintf("默认TTL测试 - %s", provider)

	fmt.Printf("使用默认TTL设置字符串: %s = %s\n", defaultTTLKey, defaultTTLValue)
	err = storageService.SetStringWithDefaultTTL(defaultTTLKey, defaultTTLValue)
	if err != nil {
		fmt.Printf("使用默认TTL设置字符串失败: %v\n", err)
		return
	}

	// 获取默认TTL字符串
	retrievedDefaultTTL, err := storageService.GetString(defaultTTLKey)
	if err != nil {
		fmt.Printf("获取默认TTL字符串失败: %v\n", err)
		return
	}

	fmt.Printf("获取默认TTL字符串: %s\n", retrievedDefaultTTL)

	// 清理测试数据
	storageService.Delete(stringKey)
	storageService.Delete(defaultTTLKey)
}

// demonstrateMultipleProviders 演示多个存储提供者
func demonstrateMultipleProviders() {
	fmt.Println("3. 多存储提供者对比演示")
	fmt.Println("========================")

	providers := []struct {
		name   string
		driver string
		config *config.StorageConfig
	}{
		{
			name:   "内存存储",
			driver: "memory",
			config: &config.StorageConfig{Driver: "memory"},
		},
		{
			name:   "BBolt存储",
			driver: "bbolt",
			config: &config.StorageConfig{
				Driver:   "bbolt",
				Database: "test.db",
			},
		},
	}

	for _, provider := range providers {
		fmt.Printf("\n--- %s ---\n", provider.name)

		storageService, err := services.NewStorageService(provider.config, nil)
		if err != nil {
			fmt.Printf("创建%s存储服务失败: %v\n", provider.name, err)
			continue
		}

		// 简单测试
		testKey := fmt.Sprintf("%s-test", provider.driver)
		testValue := []byte(fmt.Sprintf("测试数据 - %s", provider.name))

		err = storageService.Set(testKey, testValue, time.Hour)
		if err != nil {
			fmt.Printf("设置失败: %v\n", err)
			storageService.Close()
			continue
		}

		retrievedValue, err := storageService.Get(testKey)
		if err != nil {
			fmt.Printf("获取失败: %v\n", err)
		} else {
			fmt.Printf("测试成功: %s\n", string(retrievedValue))
		}

		storageService.Delete(testKey)
		storageService.Close()
	}
}
