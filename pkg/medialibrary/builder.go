package medialibrary

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MediaBuilder 为媒体库提供流式/链式参数组装构建器
type MediaBuilder struct {
	ctx              context.Context
	service          *Service
	model            HasMedia
	preserveOriginal bool
	customProps      map[string]any

	// 载荷数据
	fileData []byte
	fileName string
	err      error
}

// newMediaBuilder 构造函数
func newMediaBuilder(ctx context.Context, service *Service, model HasMedia) *MediaBuilder {
	return &MediaBuilder{
		ctx:         ctx,
		service:     service,
		model:       model,
		customProps: make(map[string]any),
	}
}

// FromFile 从本地绝对或相对文件路径读取并加载载荷
func (b *MediaBuilder) FromFile(path string) *MediaBuilder {
	if b.err != nil {
		return b
	}
	// #nosec G304
	data, err := os.ReadFile(path)
	if err != nil {
		b.err = fmt.Errorf("failed to read file: %w", err)
		return b
	}
	b.fileData = data
	b.fileName = filepath.Base(path)
	return b
}

// FromBytes 从已有的二进制字节载荷中加载
func (b *MediaBuilder) FromBytes(data []byte, filename string) *MediaBuilder {
	if b.err != nil {
		return b
	}
	b.fileData = data
	b.fileName = filename
	return b
}

// FromReader 从通用的数据读取流 io.Reader 中读取并加载
func (b *MediaBuilder) FromReader(reader io.Reader, filename string) *MediaBuilder {
	if b.err != nil {
		return b
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		b.err = fmt.Errorf("failed to read from reader: %w", err)
		return b
	}
	b.fileData = data
	b.fileName = filename
	return b
}

// PreservingOriginal 保留选项
func (b *MediaBuilder) PreservingOriginal() *MediaBuilder {
	b.preserveOriginal = true
	return b
}

// WithCustomProperties 挂载自定元数据 JSON
func (b *MediaBuilder) WithCustomProperties(props map[string]any) *MediaBuilder {
	if b.err != nil {
		return b
	}
	b.customProps = props
	return b
}

// ToMediaCollection 指定集合名称，并执行校验、转换、物理存储和数据库插入
func (b *MediaBuilder) ToMediaCollection(collectionName string) (*Media, error) {
	if b.err != nil {
		return nil, b.err
	}
	if len(b.fileData) == 0 {
		return nil, errors.New("no file content provided")
	}

	// 1. 获取模型声明的 Collections 规则
	reg := NewCollectionRegistry()
	b.model.RegisterMediaCollections(reg)
	col, hasCol := reg.Get(collectionName)

	// 2. 检测文件类型
	fileSize := int64(len(b.fileData))
	detectSize := 512
	if len(b.fileData) < 512 {
		detectSize = len(b.fileData)
	}
	mimeType := http.DetectContentType(b.fileData[:detectSize])
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = mimeType[:idx]
	}

	// 3. 执行 Collection 集合约束校验
	if hasCol {
		if col.MaxFileSize > 0 && fileSize > col.MaxFileSize {
			return nil, fmt.Errorf("file size %d exceeds collection max size limit %d", fileSize, col.MaxFileSize)
		}
		if len(col.AllowedMimes) > 0 {
			allowed := false
			for _, m := range col.AllowedMimes {
				if strings.EqualFold(m, mimeType) || strings.HasPrefix(mimeType, strings.TrimSuffix(m, "/*")) {
					allowed = true
					break
				}
			}
			if !allowed {
				return nil, fmt.Errorf("mime type %s is not allowed in collection %s", mimeType, collectionName)
			}
		}
	}

	// 4. 生成全局唯一 UUID 与定位存储
	mediaUUID := uuid.New().String()
	diskName := b.service.defaultDiskName
	disk := b.service.storage.Disk(diskName)

	cleanFileName := sanitizeFileName(b.fileName)
	originalPath := fmt.Sprintf("media/%s/%s", mediaUUID, cleanFileName)

	// 写入原图
	err := disk.Put(originalPath, b.fileData)
	if err != nil {
		return nil, fmt.Errorf("failed to save original file to disk: %w", err)
	}

	// 5. 开展 Image Conversions 图片格式转换处理
	manipulations := make(map[string]any)
	if hasCol && len(col.Conversions) > 0 && strings.HasPrefix(mimeType, "image/") {
		for _, conv := range col.Conversions {
			convData, actualFormat, err := performImageConversion(b.fileData, conv)
			if err != nil {
				// 转换错误不阻断原图上传
				continue
			}

			ext := filepath.Ext(cleanFileName)
			base := strings.TrimSuffix(cleanFileName, ext)
			targetExt := "." + actualFormat
			convFileName := fmt.Sprintf("%s-%s%s", base, conv.Name, targetExt)

			convPath := fmt.Sprintf("media/%s/conversions/%s", mediaUUID, convFileName)

			err = disk.Put(convPath, convData)
			if err != nil {
				continue
			}

			manipulations[conv.Name] = map[string]any{
				"path": convPath,
				"size": len(convData),
				"mime": "image/" + actualFormat,
			}
		}
	}

	// 6. 若为 SingleFile 覆盖模式，先移除对应集合下所有旧文件
	if hasCol && col.SingleFile {
		_ = b.service.ClearMediaCollection(b.ctx, b.model, collectionName)
	}

	// 7. 持久化入库
	baseName := strings.TrimSuffix(cleanFileName, filepath.Ext(cleanFileName))
	media := &Media{
		ModelType:        b.model.GetModelType(),
		ModelID:          b.model.GetModelID(),
		UUID:             mediaUUID,
		CollectionName:   collectionName,
		Name:             baseName,
		FileName:         cleanFileName,
		MimeType:         mimeType,
		Disk:             diskName,
		Size:             fileSize,
		Manipulations:    manipulations,
		CustomProperties: b.customProps,
		ResponsiveImages: make(map[string]any),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err = b.service.db.NewInsert().Model(media).Exec(b.ctx)
	if err != nil {
		_ = disk.DeleteDirectory(fmt.Sprintf("media/%s", mediaUUID))
		return nil, fmt.Errorf("failed to insert media record: %w", err)
	}

	return media, nil
}

// sanitizeFileName 清理文件名防跨目录攻击
func sanitizeFileName(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
