package medialibrary

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const defaultReaderLimit = 32 << 20

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
	data, err := io.ReadAll(io.LimitReader(reader, defaultReaderLimit+1))
	if err != nil {
		b.err = fmt.Errorf("failed to read from reader: %w", err)
		return b
	}
	if len(data) > defaultReaderLimit {
		b.err = fmt.Errorf("%w: reader exceeded %d bytes", ErrFileTooLarge, defaultReaderLimit)
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
		return nil, ErrNoFileContent
	}

	reg := NewCollectionRegistry()
	b.model.RegisterMediaCollections(reg)
	col, hasCol := reg.Get(collectionName)

	fileSize := int64(len(b.fileData))
	detectSize := 512
	if len(b.fileData) < 512 {
		detectSize = len(b.fileData)
	}
	mimeType := http.DetectContentType(b.fileData[:detectSize])
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = mimeType[:idx]
	}

	if err := NewCollectionPolicy(col).Validate(fileSize, mimeType, collectionName); err != nil {
		return nil, err
	}

	mediaUUID := uuid.New().String()
	diskName := b.service.defaultDiskName
	disk := b.service.storage.Disk(diskName)

	cleanFileName, err := b.service.fileNamePolicy.Sanitize(b.fileName)
	if err != nil {
		return nil, err
	}
	originalPath := b.service.pathGenerator.OriginalPath(mediaUUID, cleanFileName)

	if err := disk.Put(originalPath, b.fileData); err != nil {
		return nil, fmt.Errorf("%w: save original file: %v", ErrDiskWriteFailed, err)
	}

	conversionStatus := DerivedMediaStatusCompleted
	manipulations := make(map[string]any)
	if hasCol && len(col.Conversions) > 0 && strings.HasPrefix(mimeType, "image/") {
		conversionStatus = DerivedMediaStatusPending
		manipulations = pendingManipulations(col.Conversions)
	}

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
		ConversionStatus: conversionStatus,
		Manipulations:    manipulations,
		CustomProperties: b.customProps,
		ResponsiveImages: make(map[string]any),
		CreatedAt:        now(),
		UpdatedAt:        now(),
	}

	var oldMedia []*Media
	db, err := b.service.database()
	if err != nil {
		_ = disk.DeleteDirectory(b.service.pathGenerator.MediaDirectory(mediaUUID))
		return nil, err
	}
	err = db.RunInTx(b.ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().Model(media).Exec(ctx); err != nil {
			return fmt.Errorf("%w: %v", ErrRecordCreateFailed, err)
		}
		if hasCol && col.SingleFile {
			q := tx.NewSelect().Model(&oldMedia).
				Where("model_type = ?", b.model.GetModelType()).
				Where("model_id = ?", b.model.GetModelID()).
				Where("collection_name = ?", collectionName).
				Where("id <> ?", media.ID)
			if err := q.Scan(ctx); err != nil {
				return err
			}
			if len(oldMedia) > 0 {
				_, err := tx.NewDelete().Model((*Media)(nil)).
					Where("model_type = ?", b.model.GetModelType()).
					Where("model_id = ?", b.model.GetModelID()).
					Where("collection_name = ?", collectionName).
					Where("id <> ?", media.ID).
					Exec(ctx)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		_ = disk.DeleteDirectory(b.service.pathGenerator.MediaDirectory(mediaUUID))
		return nil, err
	}

	for _, old := range oldMedia {
		_ = disk.DeleteDirectory(b.service.pathGenerator.MediaDirectory(old.UUID))
	}

	if hasCol && len(col.Conversions) > 0 && strings.HasPrefix(mimeType, "image/") {
		if b.service.conversionMode == ConversionModeQueue {
			if err := wrapJobEnqueueError(enqueueConversionJob(b.service.queue, media.ID, col.Conversions, b.service.conversionQueue)); err != nil {
				media.ConversionStatus = DerivedMediaStatusFailed
				return media, b.service.updateConversionState(b.ctx, media, nil, DerivedMediaStatusFailed)
			}
			return media, nil
		}
		results, status, _ := NewConversionRunner(b.service).Generate(b.ctx, media, col.Conversions)
		media.Manipulations = mergeManipulations(media.Manipulations, results)
		media.ConversionStatus = status
		if err := b.service.updateConversionState(b.ctx, media, media.Manipulations, status); err != nil {
			return media, err
		}
	}

	return media, nil
}
