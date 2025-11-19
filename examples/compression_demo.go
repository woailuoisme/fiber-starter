package main

import (
	"fmt"
	"log"
	"time"

	"fiber-starter/app/services"
	"fiber-starter/config"
)

func main() {
	fmt.Println("=== 存储压缩功能演示 ===")

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

	// 测试数据
	testData := []byte("这是一个测试字符串，用于演示压缩功能。这个字符串包含了一些重复的内容，以便更好地展示压缩效果。压缩功能可以帮助减少存储空间的使用，特别是在存储大量重复数据时效果更加明显。")

	fmt.Printf("原始数据大小: %d 字节\n", len(testData))
	fmt.Printf("原始数据内容: %s\n\n", string(testData))

	// 测试无压缩
	fmt.Println("=== 测试无压缩 ===")
	storageService.SetCompression(services.CompressionNone)

	err = storageService.Set("test_no_compress", testData, time.Minute*10)
	if err != nil {
		log.Printf("设置无压缩数据失败: %v", err)
	} else {
		retrieved, err := storageService.Get("test_no_compress")
		if err != nil {
			log.Printf("获取无压缩数据失败: %v", err)
		} else {
			fmt.Printf("获取数据大小: %d 字节\n", len(retrieved))
			fmt.Printf("数据内容: %s\n", string(retrieved))
			fmt.Printf("压缩率: %.2f%%\n\n", float64(len(testData)-len(retrieved))/float64(len(testData))*100)
		}
	}

	// 测试Gzip压缩
	fmt.Println("=== 测试Gzip压缩 ===")
	err = storageService.SetCompression(services.CompressionGzip)
	if err != nil {
		log.Printf("设置Gzip压缩失败: %v", err)
	} else {
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
				fmt.Printf("数据内容: %s\n", string(retrieved))
				fmt.Printf("压缩率: %.2f%%\n\n", float64(len(testData)-len(retrieved))/float64(len(testData))*100)
			}
		}
	}

	// 测试Zstd压缩
	fmt.Println("=== 测试Zstd压缩 ===")
	err = storageService.SetCompression(services.CompressionZstd)
	if err != nil {
		log.Printf("设置Zstd压缩失败: %v", err)
	} else {
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
				fmt.Printf("数据内容: %s\n", string(retrieved))
				fmt.Printf("压缩率: %.2f%%\n\n", float64(len(testData)-len(retrieved))/float64(len(testData))*100)
			}
		}
	}

	// 测试不同大小数据的压缩效果
	fmt.Println("=== 测试不同大小数据的压缩效果 ===")
	testSizes := []int{100, 500, 1000, 5000}

	for _, size := range testSizes {
		// 生成重复数据以获得更好的压缩效果
		data := make([]byte, size)
		pattern := []byte("重复数据模式ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
		for i := 0; i < size; i++ {
			data[i] = pattern[i%len(pattern)]
		}

		fmt.Printf("\n数据大小: %d 字节\n", size)

		// 测试Gzip
		storageService.SetCompression(services.CompressionGzip)
		storageService.Set("gzip_test", data, time.Minute)
		gzipData, _ := storageService.Get("gzip_test")
		gzipRatio := float64(size-len(gzipData)) / float64(size) * 100

		// 测试Zstd
		storageService.SetCompression(services.CompressionZstd)
		storageService.Set("zstd_test", data, time.Minute)
		zstdData, _ := storageService.Get("zstd_test")
		zstdRatio := float64(size-len(zstdData)) / float64(size) * 100

		fmt.Printf("Gzip压缩后: %d 字节 (压缩率: %.2f%%)\n", len(gzipData), gzipRatio)
		fmt.Printf("Zstd压缩后: %d 字节 (压缩率: %.2f%%)\n", len(zstdData), zstdRatio)
		fmt.Printf("Zstd相对Gzip的改进: %.2f%%\n", zstdRatio-gzipRatio)
	}

	// 清理测试数据
	fmt.Println("\n=== 清理测试数据 ===")
	keys := []string{"test_no_compress", "test_gzip", "test_zstd", "gzip_test", "zstd_test"}
	for _, key := range keys {
		err := storageService.Delete(key)
		if err != nil {
			log.Printf("删除键 %s 失败: %v", key, err)
		}
	}

	fmt.Println("压缩功能演示完成！")
}
