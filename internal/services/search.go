package services

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"fiber-starter/internal/config"

	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
)

// SearchService 搜索服务接口
type SearchService interface {
	// CreateIndex 索引操作
	CreateIndex(uid string, primaryKey string) (*meilisearch.TaskInfo, error)
	GetIndex(uid string) (meilisearch.IndexManager, error)
	DeleteIndex(uid string) (*meilisearch.TaskInfo, error)

	// AddDocuments 文档操作
	AddDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error)
	UpdateDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error)
	DeleteDocuments(indexUID string, ids []string) (*meilisearch.TaskInfo, error)
	DeleteAllDocuments(indexUID string) (*meilisearch.TaskInfo, error)

	// Search 搜索操作
	Search(indexUID string, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)

	// Health 健康检查
	Health() (bool, error)
}

// searchService 搜索服务实现
type searchService struct {
	client      meilisearch.ServiceManager
	logger      *zap.Logger
	config      *config.Config
	initMu      sync.Mutex
	initialized uint32 // 0: false, 1: true
}

// NewSearchService 创建新的搜索服务实例
func NewSearchService(cfg *config.Config, logger *zap.Logger) (SearchService, error) {
	// 延迟初始化
	return &searchService{
		config: cfg,
		logger: logger,
	}, nil
}

// ensureInitialized 确保搜索服务已初始化
func (s *searchService) ensureInitialized() error {
	if atomic.LoadUint32(&s.initialized) == 1 {
		return nil
	}

	s.initMu.Lock()
	defer s.initMu.Unlock()

	if s.initialized == 1 {
		return nil
	}

	if s.config.Meilisearch.Host == "" {
		return fmt.Errorf("meilisearch host is required")
	}

	client := meilisearch.New(s.config.Meilisearch.Host,
		meilisearch.WithAPIKey(s.config.Meilisearch.APIKey),
		meilisearch.WithCustomClient(&http.Client{
			Timeout: 5 * time.Second,
		}),
	)

	// 验证连接
	// 注意：Health 检查可能因为 Meilisearch 服务未启动而失败，但不应阻止服务启动
	// 实际生产环境中可能需要更健壮的处理
	if _, err := client.Health(); err != nil {
		s.logger.Warn("Failed to connect to Meilisearch", zap.Error(err))
	} else {
		s.logger.Info("Connected to Meilisearch", zap.String("host", s.config.Meilisearch.Host))
	}

	s.client = client
	atomic.StoreUint32(&s.initialized, 1)
	return nil
}

func (s *searchService) withClient(fn func(meilisearch.ServiceManager) error) error {
	if err := s.ensureInitialized(); err != nil {
		return err
	}
	return fn(s.client)
}

// CreateIndex 创建索引
func (s *searchService) CreateIndex(uid string, primaryKey string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.CreateIndex(&meilisearch.IndexConfig{Uid: uid, PrimaryKey: primaryKey})
		return err
	})
	return info, err
}

// GetIndex 获取索引
func (s *searchService) GetIndex(uid string) (meilisearch.IndexManager, error) {
	var index meilisearch.IndexManager
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		index = client.Index(uid)
		return nil
	})
	return index, err
}

// DeleteIndex 删除索引
func (s *searchService) DeleteIndex(uid string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.DeleteIndex(uid)
		return err
	})
	return info, err
}

// AddDocuments 添加文档
func (s *searchService) AddDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).AddDocuments(documents, nil)
		return err
	})
	return info, err
}

// UpdateDocuments 更新文档
func (s *searchService) UpdateDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).UpdateDocuments(documents, nil)
		return err
	})
	return info, err
}

// DeleteDocuments 删除文档
func (s *searchService) DeleteDocuments(indexUID string, ids []string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).DeleteDocuments(ids, nil)
		return err
	})
	return info, err
}

// DeleteAllDocuments 删除所有文档
func (s *searchService) DeleteAllDocuments(indexUID string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).DeleteAllDocuments(nil)
		return err
	})
	return info, err
}

// Search 搜索
func (s *searchService) Search(indexUID string, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	var response *meilisearch.SearchResponse
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		response, err = client.Index(indexUID).Search(query, request)
		return err
	})
	return response, err
}

// Health 健康检查
func (s *searchService) Health() (bool, error) {
	var healthy bool
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		resp, err := client.Health()
		if err != nil {
			return err
		}
		healthy = resp.Status == "available"
		return nil
	})
	return healthy, err
}
