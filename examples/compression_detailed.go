package main

import (
	"fmt"
	"log"
	"time"

	"fiber-starter/app/services"
	"fiber-starter/config"
)

func main() {
	fmt.Println("=== 存储压缩功能详细测试 ===")

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
	testData := []byte("这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。这是一个测试字符串，用于演示压缩功能。")

	fmt.Printf("原始数据大小: %d 字节\n\n", len(testData))

	// 测试压缩后直接存储
	fmt.Println("=== 测试压缩后直接存储 ===")

	// 设置为Gzip压缩
	err = storageService.SetCompression(services.CompressionGzip)
	if err != nil {
		log.Printf("设置Gzip压缩失败: %v", err)
	}

	// 先压缩数据
	compressedData, err := storageService.CompressData(testData)
	if err != nil {
		log.Printf("压缩数据失败: %v", err)
	} else {
		fmt.Printf("压缩后数据大小: %d 字节 (压缩率: %.2f%%)\n", len(compressedData), float64(len(testData)-len(compressedData))/float64(len(testData))*100)

		// 直接存储压缩数据
		err = storageService.GetStorage().Set("compressed_direct", compressedData, time.Minute)
		if err != nil {
			log.Printf("直接存储压缩数据失败: %v", err)
		} else {
			// 获取数据
			retrieved, err := storageService.GetStorage().Get("compressed_direct")
			if err != nil {
				log.Printf("获取压缩数据失败: %v", err)
			} else {
				fmt.Printf("获取压缩数据大小: %d 字节\n", len(retrieved))

				// 解压数据
				decompressed, err := storageService.DecompressData(retrieved)
				if err != nil {
					log.Printf("解压数据失败: %v", err)
				} else {
					fmt.Printf("解压后数据大小: %d 字节\n", len(decompressed))
					fmt.Printf("数据匹配: %t\n", string(testData) == string(decompressed))
				}
			}
		}
	}

	// 测试通过Set方法的压缩
	fmt.Println("\n=== 测试通过Set方法的压缩 ===")

	// 设置为Zstd压缩
	err = storageService.SetCompression(services.CompressionZstd)
	if err != nil {
		log.Printf("设置Zstd压缩失败: %v", err)
	}

	// 通过Set方法存储（应该自动压缩）
	err = storageService.Set("compressed_auto", testData, time.Minute)
	if err != nil {
		log.Printf("自动压缩存储失败: %v", err)
	} else {
		// 直接从底层存储获取（不通过解压）
		retrieved, err := storageService.GetStorage().Get("compressed_auto")
		if err != nil {
			log.Printf("获取存储数据失败: %v", err)
		} else {
			fmt.Printf("底层存储数据大小: %d 字节\n", len(retrieved))
			fmt.Printf("是否被压缩: %t\n", len(retrieved) < len(testData))
		}

		// 通过Get方法获取（应该自动解压）
		decompressed, err := storageService.Get("compressed_auto")
		if err != nil {
			log.Printf("获取解压数据失败: %v", err)
		} else {
			fmt.Printf("通过Get获取数据大小: %d 字节\n", len(decompressed))
			fmt.Printf("数据匹配: %t\n", string(testData) == string(decompressed))
		}
	}

	// 测试不同压缩类型的效果对比
	fmt.Println("\n=== 压缩类型效果对比 ===")

	compressionTypes := []struct {
		name string
		typ  services.CompressionType
	}{
		{"无压缩", services.CompressionNone},
		{"Gzip", services.CompressionGzip},
		{"Zstd", services.CompressionZstd},
	}

	for _, ct := range compressionTypes {
		err = storageService.SetCompression(ct.typ)
		if err != nil {
			log.Printf("设置压缩类型 %s 失败: %v", ct.name, err)
			continue
		}

		key := fmt.Sprintf("test_%s", ct.name)
		err = storageService.Set(key, testData, time.Minute)
		if err != nil {
			log.Printf("存储数据失败 (%s): %v", ct.name, err)
			continue
		}

		// 获取底层存储的数据大小
		rawData, err := storageService.GetStorage().Get(key)
		if err != nil {
			log.Printf("获取原始数据失败 (%s): %v", ct.name, err)
			continue
		}

		// 通过Get方法获取解压后的数据
		finalData, err := storageService.Get(key)
		if err != nil {
			log.Printf("获取最终数据失败 (%s): %v", ct.name, err)
			continue
		}

		compressionRatio := float64(len(testData)-len(rawData)) / float64(len(testData)) * 100
		fmt.Printf("%s: 存储%d字节, 获取%d字节, 压缩率: %.2f%%, 数据正确: %t\n",
			ct.name, len(rawData), len(finalData), compressionRatio, string(testData) == string(finalData))
	}

	// 清理测试数据
	fmt.Println("\n=== 清理测试数据 ===")
	keys := []string{"compressed_direct", "compressed_auto", "test_无压缩", "test_Gzip", "test_Zstd"}
	for _, key := range keys {
		err := storageService.Delete(key)
		if err != nil {
			log.Printf("删除键 %s 失败: %v", key, err)
		}
	}

	fmt.Println("详细测试完成！")
}
