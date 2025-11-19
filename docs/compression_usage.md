# 存储压缩功能使用说明

## 概述

存储服务现在支持数据压缩功能，可以使用 Gzip 或 Zstd 算法自动压缩存储的数据，从而节省存储空间。

## 支持的压缩类型

1. **无压缩** (`CompressionNone`) - 不进行任何压缩
2. **Gzip压缩** (`CompressionGzip`) - 使用标准Gzip算法
3. **Zstd压缩** (`CompressionZstd`) - 使用Zstandard算法（推荐，压缩率更高）

## 使用方法

### 1. 基本使用

```go
import (
    "fiber-starter/app/services"
    "fiber-starter/config"
)

// 创建存储服务
storageConfig := &config.StorageConfig{
    Driver: "memory", // 或其他驱动如 "redis", "bbolt", "minio", "s3"
}

storageService, err := services.NewStorageService(storageConfig, nil)
if err != nil {
    log.Fatal(err)
}
defer storageService.Close()

// 设置压缩类型
err = storageService.SetCompression(services.CompressionZstd)
if err != nil {
    log.Fatal(err)
}

// 存储数据（会自动压缩）
data := []byte("要压缩存储的数据...")
err = storageService.Set("my_key", data, time.Hour)
if err != nil {
    log.Fatal(err)
}

// 获取数据（会自动解压）
retrieved, err := storageService.Get("my_key")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("原始数据: %s\n", string(retrieved))
```

### 2. 压缩类型切换

```go
// 切换到Gzip压缩
storageService.SetCompression(services.CompressionGzip)

// 切换到Zstd压缩
storageService.SetCompression(services.CompressionZstd)

// 禁用压缩
storageService.SetCompression(services.CompressionNone)
```

### 3. 获取压缩信息

```go
// 获取当前压缩类型
compressionType := storageService.GetCompressionType()

// 获取压缩类型名称
typeName := storageService.GetCompressionTypeName()
fmt.Printf("当前压缩类型: %s\n", typeName)
```

### 4. 手动压缩/解压

```go
// 手动压缩数据
compressed, err := storageService.CompressData(originalData)
if err != nil {
    log.Fatal(err)
}

// 手动解压数据
decompressed, err := storageService.DecompressData(compressed)
if err != nil {
    log.Fatal(err)
}
```

## 性能对比

基于测试结果（570字节的重复数据）：

| 压缩类型 | 压缩后大小 | 压缩率 | 适用场景 |
|---------|-----------|--------|----------|
| 无压缩 | 570字节 | 0% | 小数据或实时性要求高 |
| Gzip | 92字节 | 83.86% | 兼容性好，通用场景 |
| Zstd | 82字节 | 85.61% | 推荐使用，压缩率最高 |

## 使用建议

### 何时使用压缩

1. **大数据存储**：当存储的数据量较大时，压缩可以显著节省空间
2. **重复数据**：包含大量重复内容的数据压缩效果更好
3. **文本数据**：JSON、XML、日志等文本数据压缩效果明显
4. **成本敏感**：当存储成本是重要考虑因素时

### 何时避免压缩

1. **小数据**：小于100字节的数据压缩效果不明显
2. **已压缩数据**：如图片、视频等已压缩的数据
3. **实时性要求高**：压缩/解压会增加CPU开销
4. **频繁访问**：频繁读写的数据会增加CPU负担

### 压缩类型选择

- **推荐Zstd**：压缩率最高，性能好，适合大多数场景
- **Gzip**：兼容性更好，适合需要与其他系统交互的场景
- **无压缩**：小数据或性能敏感的场景

## 注意事项

1. **透明性**：压缩对用户是透明的，Set方法自动压缩，Get方法自动解压
2. **CPU开销**：压缩会增加CPU使用，需要在空间和性能间权衡
3. **内存使用**：压缩过程需要额外的内存空间
4. **数据一致性**：压缩不会影响数据完整性，所有压缩数据都能正确解压
5. **存储驱动兼容**：压缩功能与所有存储驱动兼容（内存、Redis、BBolt、MinIO、S3）

## 示例代码

完整的使用示例请参考：
- `examples/compression_demo.go` - 基本功能演示
- `examples/compression_debug.go` - 调试版本
- `examples/compression_detailed.go` - 详细测试

## 技术实现

压缩功能通过以下方式实现：

1. **Set方法**：在存储前根据设置的压缩类型压缩数据
2. **Get方法**：在获取后根据设置的压缩类型解压数据
3. **压缩算法**：
   - Gzip：使用标准库 `compress/gzip`
   - Zstd：使用 `github.com/klauspost/compress/zstd`
4. **错误处理**：压缩/解压失败会返回详细的错误信息

## 依赖包

压缩功能需要以下依赖：

```bash
go get github.com/klauspost/compress
```

该包提供了高性能的Zstd压缩算法实现。