package medialibrary

import (
	"context"
	"strings"
	"time"

	databaseContracts "lfiber/internal/providers/database/contracts"
	queueContracts "lfiber/internal/providers/queue/contracts"
	storageContracts "lfiber/internal/providers/storage/contracts"

	"github.com/uptrace/bun"
)

// Service 媒体库核心业务逻辑服务
type Service struct {
	db              bun.IDB
	dbResolver      func() (bun.IDB, error)
	storage         storageContracts.StorageManager
	defaultDiskName string
	conversionMode  string
	conversionQueue string
	queue           queueContracts.Queue
	pathGenerator   PathGenerator
	fileNamePolicy  FileNamePolicy
}

type RegenerateOptions struct {
	OnlyMissing bool
	FailedOnly  bool
	Collection  string
}

type Option func(*Service)

func WithConversionMode(mode string) Option {
	return func(s *Service) {
		mode = strings.ToLower(strings.TrimSpace(mode))
		if mode == ConversionModeQueue {
			s.conversionMode = ConversionModeQueue
			return
		}
		s.conversionMode = ConversionModeSync
	}
}

func WithQueue(queue queueContracts.Queue, queueName string) Option {
	return func(s *Service) {
		s.queue = queue
		s.conversionQueue = queueName
	}
}

// NewService 实例化媒体库服务
func NewService(db bun.IDB, storage storageContracts.StorageManager, defaultDiskName string, opts ...Option) *Service {
	s := &Service{
		db:              db,
		dbResolver:      func() (bun.IDB, error) { return db, nil },
		storage:         storage,
		defaultDiskName: defaultDiskName,
		conversionMode:  ConversionModeSync,
		conversionQueue: "default",
		pathGenerator:   PathGenerator{},
		fileNamePolicy:  NewFileNamePolicy(),
	}
	for _, opt := range opts {
		opt(s)
	}
	setDefaultService(s)
	return s
}

func NewServiceFromConnection(connection databaseContracts.Connection, storage storageContracts.StorageManager, defaultDiskName string, opts ...Option) *Service {
	s := &Service{
		dbResolver: func() (bun.IDB, error) {
			return connection.BunDB()
		},
		storage:         storage,
		defaultDiskName: defaultDiskName,
		conversionMode:  ConversionModeSync,
		conversionQueue: "default",
		pathGenerator:   PathGenerator{},
		fileNamePolicy:  NewFileNamePolicy(),
	}
	for _, opt := range opts {
		opt(s)
	}
	setDefaultService(s)
	return s
}

func (s *Service) database() (bun.IDB, error) {
	if s.db != nil {
		return s.db, nil
	}
	return s.dbResolver()
}

// AddMedia 启动链式构建器以关联上传新媒体文件
func (s *Service) AddMedia(ctx context.Context, model HasMedia) *MediaBuilder {
	return newMediaBuilder(ctx, s, model)
}

// GetMedia 检索关联到指定模型的媒体记录
func (s *Service) GetMedia(ctx context.Context, model HasMedia, collection ...string) ([]*Media, error) {
	var list []*Media
	db, err := s.database()
	if err != nil {
		return nil, err
	}
	q := db.NewSelect().Model(&list).
		Where("model_type = ?", model.GetModelType()).
		Where("model_id = ?", model.GetModelID()).
		Order("order_column ASC", "id ASC")

	if len(collection) > 0 && collection[0] != "" {
		q.Where("collection_name = ?", collection[0])
	}

	err = q.Scan(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// DeleteMedia 物理移除媒体关联的所有磁盘文件，并抹除数据库记录
func (s *Service) DeleteMedia(ctx context.Context, media *Media) error {
	disk := s.storage.Disk(media.Disk)
	dir := s.pathGenerator.MediaDirectory(media.UUID)

	// 物理级联删除目录下的所有内容（原图与 conversions）
	_ = disk.DeleteDirectory(dir)

	// 清理数据库记录
	db, err := s.database()
	if err != nil {
		return err
	}
	_, err = db.NewDelete().Model(media).Where("id = ?", media.ID).Exec(ctx)
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
		if p, ok := conversionPath(media.Manipulations, convName); ok {
			return p
		}
	}
	// 默认退避返回原图物理路径
	return s.pathGenerator.OriginalPath(media.UUID, media.FileName)
}

func (s *Service) RunConversions(ctx context.Context, mediaID int64, rules []*ConversionRule) error {
	var media Media
	db, err := s.database()
	if err != nil {
		return err
	}
	if err := db.NewSelect().Model(&media).Where("id = ?", mediaID).Scan(ctx); err != nil {
		return err
	}
	results, status, err := NewConversionRunner(s).Generate(ctx, &media, rules)
	media.Manipulations = mergeManipulations(media.Manipulations, results)
	media.ConversionStatus = status
	media.UpdatedAt = now()
	_, updateErr := db.NewUpdate().
		Model(&media).
		Column("manipulations", "conversion_status", "updated_at").
		Where("id = ?", media.ID).
		Exec(ctx)
	if updateErr != nil {
		return updateErr
	}
	return err
}

func (s *Service) Regenerate(ctx context.Context, opts RegenerateOptions) (int, error) {
	var list []*Media
	db, err := s.database()
	if err != nil {
		return 0, err
	}
	q := db.NewSelect().Model(&list).Order("id ASC")
	if opts.Collection != "" {
		q.Where("collection_name = ?", opts.Collection)
	}
	if opts.FailedOnly {
		q.Where("conversion_status = ?", DerivedMediaStatusFailed)
	}
	if err := q.Scan(ctx); err != nil {
		return 0, err
	}

	count := 0
	for _, media := range list {
		rules := conversionRulesFromManipulations(media.Manipulations, opts.OnlyMissing)
		if len(rules) == 0 {
			continue
		}
		if err := s.RunConversions(ctx, media.ID, rules); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s *Service) updateConversionState(ctx context.Context, media *Media, manipulations map[string]any, status string) error {
	if manipulations != nil {
		media.Manipulations = manipulations
	}
	media.ConversionStatus = status
	media.UpdatedAt = now()
	db, err := s.database()
	if err != nil {
		return err
	}
	_, err = db.NewUpdate().
		Model(media).
		Column("manipulations", "conversion_status", "updated_at").
		Where("id = ?", media.ID).
		Exec(ctx)
	return err
}

func conversionPath(manipulations map[string]any, conversionName string) (string, bool) {
	if manipulations == nil {
		return "", false
	}
	switch conv := manipulations[conversionName].(type) {
	case map[string]any:
		if status, _ := conv["status"].(string); status != "" && status != DerivedMediaStatusCompleted {
			return "", false
		}
		if p, ok := conv["path"].(string); ok && p != "" {
			return p, true
		}
	case ConversionResult:
		if conv.Status == DerivedMediaStatusCompleted && conv.Path != "" {
			return conv.Path, true
		}
	}
	return "", false
}

func mergeManipulations(current, next map[string]any) map[string]any {
	if current == nil {
		current = make(map[string]any)
	}
	for key, value := range next {
		switch v := value.(type) {
		case ConversionResult:
			existing, _ := current[key].(map[string]any)
			if existing == nil {
				existing = make(map[string]any)
			}
			existing["path"] = v.Path
			existing["size"] = v.Size
			existing["mime_type"] = v.MimeType
			existing["mime"] = v.MimeType
			existing["status"] = v.Status
			current[key] = existing
		default:
			current[key] = value
		}
	}
	return current
}

func pendingManipulations(rules []*ConversionRule) map[string]any {
	manipulations := make(map[string]any, len(rules))
	for _, rule := range rules {
		manipulations[rule.Name] = map[string]any{
			"width":  rule.Width,
			"height": rule.Height,
			"crop":   rule.Crop,
			"format": rule.Format,
			"status": DerivedMediaStatusPending,
		}
	}
	return manipulations
}

func conversionRulesFromManipulations(manipulations map[string]any, onlyMissing bool) []*ConversionRule {
	if len(manipulations) == 0 {
		return nil
	}
	rules := make([]*ConversionRule, 0, len(manipulations))
	for name, value := range manipulations {
		item, ok := value.(map[string]any)
		if !ok {
			continue
		}
		if onlyMissing {
			status, _ := item["status"].(string)
			path, _ := item["path"].(string)
			if status == DerivedMediaStatusCompleted && path != "" {
				continue
			}
		}
		rules = append(rules, &ConversionRule{
			Name:   name,
			Width:  intFromAny(item["width"]),
			Height: intFromAny(item["height"]),
			Crop:   boolFromAny(item["crop"]),
			Format: stringFromAny(item["format"]),
		})
	}
	return rules
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return 0
	}
}

func boolFromAny(value any) bool {
	v, _ := value.(bool)
	return v
}

func stringFromAny(value any) string {
	v, _ := value.(string)
	return v
}
