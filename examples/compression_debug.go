package main

import (
	"fmt"
	"log"
	"time"

	"fiber-starter/app/services"
	"fiber-starter/config"
)

func main() {
	fmt.Println("=== 存储压缩功能调试演示 ===")

	// 创建内存存储配置
	storageConfig := &config.StorageConfig{
		Driver: "memory",
	}

	// 创建存储服务
	storageService, err := services.NewStorageService(storageConfig, nil)
	if err != nil {
		log.Fatalf("创建存储服务失败: %v", err)
	}
	defer storageService.Close()

	// 测试数据 - 使用更多重复内容
	testData := []byte("这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。")

	fmt.Printf("原始数据大小: %d 字节\n", len(testData))
	fmt.Printf("原始数据内容: %s\n\n", string(testData))

	// 测试无压缩
	fmt.Println("=== 测试无压缩 ===")
	err = storageService.SetCompression(services.CompressionNone)
	if err != nil {
		log.Printf("设置无压缩失败: %v", err)
	}

	fmt.Printf("当前压缩类型: %s\n", storageService.GetCompressionTypeName())

	err = storageService.Set("test_no_compress", testData, time.Minute*10)
	if err != nil {
		log.Printf("设置无压缩数据失败: %v", err)
	} else {
		retrieved, err := storageService.Get("test_no_compress")
		if err != nil {
			log.Printf("获取无压缩数据失败: %v", err)
		} else {
			fmt.Printf("获取数据大小: %d 字节\n", len(retrieved))
			fmt.Printf("数据匹配: %t\n", string(testData) == string(retrieved))
			fmt.Printf("压缩率: %.2f%%\n\n", float64(len(testData)-len(retrieved))/float64(len(testData))*100)
		}
	}

	// 测试Gzip压缩
	fmt.Println("=== 测试Gzip压缩 ===")
	err = storageService.SetCompression(services.CompressionGzip)
	if err != nil {
		log.Printf("设置Gzip压缩失败: %v", err)
	}

	fmt.Printf("当前压缩类型: %s\n", storageService.GetCompressionTypeName())

	err = storageService.Set("test_gzip", testData, time.Minute*10)
	if err != nil {
		log.Printf("设置Gzip压缩数据失败: %v", err)
	} else {
		retrieved, err := storageService.Get("test_gzip")
		if err != nil {
			log.Printf("获取Gzip压缩数据失败: %v", err)
		} else {
			fmt.Printf("获取数据大小: %d 字节\n", len(retrieved))
			fmt.Printf("数据匹配: %t\n", string(testData) == string(retrieved))
			fmt.Printf("压缩率: %.2f%%\n\n", float64(len(testData)-len(retrieved))/float64(len(testData))*100)
		}
	}

	// 测试Zstd压缩
	fmt.Println("=== 测试Zstd压缩 ===")
	err = storageService.SetCompression(services.CompressionZstd)
	if err != nil {
		log.Printf("设置Zstd压缩失败: %v", err)
	}

	fmt.Printf("当前压缩类型: %s\n", storageService.GetCompressionTypeName())

	err = storageService.Set("test_zstd", testData, time.Minute*10)
	if err != nil {
		log.Printf("设置Zstd压缩数据失败: %v", err)
	} else {
		retrieved, err := storageService.Get("test_zstd")
		if err != nil {
			log.Printf("获取Zstd压缩数据失败: %v", err)
		} else {
			fmt.Printf("获取数据大小: %d 字节\n", len(retrieved))
			fmt.Printf("数据匹配: %t\n", string(testData) == string(retrieved))
			fmt.Printf("压缩率: %.2f%%\n\n", float64(len(testData)-len(retrieved))/float64(len(testData))*100)
		}
	}

	// 直接测试压缩函数
	fmt.Println("=== 直接测试压缩函数 ===")

	// 测试Gzip压缩
	storageService.SetCompression(services.CompressionGzip)
	compressed, err := storageService.CompressData(testData)
	if err != nil {
		log.Printf("Gzip压缩失败: %v", err)
	} else {
		fmt.Printf("Gzip压缩后大小: %d 字节 (压缩率: %.2f%%)\n", len(compressed), float64(len(testData)-len(compressed))/float64(len(testData))*100)

		decompressed, err := storageService.DecompressData(compressed)
		if err != nil {
			log.Printf("Gzip解压失败: %v", err)
		} else {
			fmt.Printf("Gzip解压后大小: %d 字节\n", len(decompressed))
			fmt.Printf("Gzip数据匹配: %t\n", string(testData) == string(decompressed))
		}
	}

	// 测试Zstd压缩
	storageService.SetCompression(services.CompressionZstd)
	compressed, err = storageService.CompressData(testData)
	if err != nil {
		log.Printf("Zstd压缩失败: %v", err)
	} else {
		fmt.Printf("Zstd压缩后大小: %d 字节 (压缩率: %.2f%%)\n", len(compressed), float64(len(testData)-len(compressed))/float64(len(testData))*100)

		decompressed, err := storageService.DecompressData(compressed)
		if err != nil {
			log.Printf("Zstd解压失败: %v", err)
		} else {
			fmt.Printf("Zstd解压后大小: %d 字节\n", len(decompressed))
			fmt.Printf("Zstd数据匹配: %t\n", string(testData) == string(decompressed))
		}
	}

	// 清理测试数据
	fmt.Println("\n=== 清理测试数据 ===")
	keys := []string{"test_no_compress", "test_gzip", "test_zstd"}
	for _, key := range keys {
		err := storageService.Delete(key)
		if err != nil {
			log.Printf("删除键 %s 失败: %v", key, err)
		}
	}

	fmt.Println("压缩功能调试演示完成！")
}
