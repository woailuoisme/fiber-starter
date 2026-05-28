package medialibrary

import (
	"context"
	"fmt"
	"time"

	storageContracts "lfiber/internal/providers/storage/contracts"

	"github.com/uptrace/bun"
)

// Service 媒体库核心业务逻辑服务
type Service struct {
	db              bun.IDB
	storage         storageContracts.StorageManager
	defaultDiskName string
}

// NewService 实例化媒体库服务
func NewService(db bun.IDB, storage storageContracts.StorageManager, defaultDiskName string) *Service {
	return &Service{
		db:              db,
		storage:         storage,
		defaultDiskName: defaultDiskName,
	}
}

// AddMedia 启动链式构建器以关联上传新媒体文件
func (s *Service) AddMedia(ctx context.Context, model HasMedia) *MediaBuilder {
	return newMediaBuilder(ctx, s, model)
}

// GetMedia 检索关联到指定模型的媒体记录
func (s *Service) GetMedia(ctx context.Context, model HasMedia, collection ...string) ([]*Media, error) {
	var list []*Media
	q := s.db.NewSelect().Model(&list).
		Where("model_type = ?", model.GetModelType()).
		Where("model_id = ?", model.GetModelID()).
		Order("order_column ASC", "id ASC")

	if len(collection) > 0 && collection[0] != "" {
		q.Where("collection_name = ?", collection[0])
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteMedia 物理移除媒体关联的所有磁盘文件，并抹除数据库记录
func (s *Service) DeleteMedia(ctx context.Context, media *Media) error {
	disk := s.storage.Disk(media.Disk)
	dir := fmt.Sprintf("media/%s", media.UUID)

	// 物理级联删除目录下的所有内容（原图与 conversions）
	_ = disk.DeleteDirectory(dir)

	// 清理数据库记录
	_, err := s.db.NewDelete().Model(media).Where("id = ?", media.ID).Exec(ctx)
	return err
}

// ClearMediaCollection 清除指定模型在特定集合下的全部媒体文件与记录
func (s *Service) ClearMediaCollection(ctx context.Context, model HasMedia, collection string) error {
	list, err := s.GetMedia(ctx, model, collection)
	if err != nil {
		return err
	}
	for _, m := range list {
		if err := s.DeleteMedia(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

// GetUrl 衍生获取指定图片版本或原图的可公开访问 URL
func (s *Service) GetUrl(media *Media, conversion ...string) string {
	disk := s.storage.Disk(media.Disk)
	path := s.GetPath(media, conversion...)
	return disk.Url(path)
}

// GetTemporaryUrl 衍生获取指定图片版本或原图的带时效安全访问 URL
func (s *Service) GetTemporaryUrl(media *Media, expiration time.Duration, conversion ...string) (string, error) {
	disk := s.storage.Disk(media.Disk)
	path := s.GetPath(media, conversion...)
	return disk.TemporaryUrl(path, expiration)
}

// GetPath 获取原图或目标转换图在磁盘上的相对路径
func (s *Service) GetPath(media *Media, conversion ...string) string {
	if len(conversion) > 0 && conversion[0] != "" {
		convName := conversion[0]
		if conv, ok := media.Manipulations[convName].(map[string]any); ok {
			if p, ok := conv["path"].(string); ok {
				return p
			}
		}
	}
	// 默认退避返回原图物理路径
	return fmt.Sprintf("media/%s/%s", media.UUID, media.FileName)
}
