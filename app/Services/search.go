package services

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	helpers "fiber-starter/app/Support"
	"fiber-starter/config"

	"github.com/meilisearch/meilisearch-go"
)

// SearchService 搜索服务接口
type SearchService interface {
	CreateIndex(uid string, primaryKey string) (*meilisearch.TaskInfo, error)
	GetIndex(uid string) (meilisearch.IndexManager, error)
	DeleteIndex(uid string) (*meilisearch.TaskInfo, error)
	AddDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error)
	UpdateDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error)
	DeleteDocuments(indexUID string, ids []string) (*meilisearch.TaskInfo, error)
	DeleteAllDocuments(indexUID string) (*meilisearch.TaskInfo, error)
	Search(indexUID string, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error)
	Health() (bool, error)
}

type searchService struct {
	client      meilisearch.ServiceManager
	config      *config.Config
	initMu      sync.Mutex
	initialized uint32
}

func NewSearchService(cfg *config.Config) SearchService {
	return &searchService{config: cfg}
}

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
		meilisearch.WithCustomClient(&http.Client{Timeout: 5 * time.Second}),
	)

	if _, err := client.Health(); err != nil {
		helpers.Warn("Failed to connect to Meilisearch")
	} else {
		helpers.Info("Connected to Meilisearch")
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

func (s *searchService) CreateIndex(uid string, primaryKey string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.CreateIndex(&meilisearch.IndexConfig{Uid: uid, PrimaryKey: primaryKey})
		return err
	})
	return info, err
}

func (s *searchService) GetIndex(uid string) (meilisearch.IndexManager, error) {
	var index meilisearch.IndexManager
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		index = client.Index(uid)
		return nil
	})
	return index, err
}

func (s *searchService) DeleteIndex(uid string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.DeleteIndex(uid)
		return err
	})
	return info, err
}

func (s *searchService) AddDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).AddDocuments(documents, nil)
		return err
	})
	return info, err
}

func (s *searchService) UpdateDocuments(indexUID string, documents interface{}) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).UpdateDocuments(documents, nil)
		return err
	})
	return info, err
}

func (s *searchService) DeleteDocuments(indexUID string, ids []string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).DeleteDocuments(ids, nil)
		return err
	})
	return info, err
}

func (s *searchService) DeleteAllDocuments(indexUID string) (*meilisearch.TaskInfo, error) {
	var info *meilisearch.TaskInfo
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		info, err = client.Index(indexUID).DeleteAllDocuments(nil)
		return err
	})
	return info, err
}

func (s *searchService) Search(indexUID string, query string, request *meilisearch.SearchRequest) (*meilisearch.SearchResponse, error) {
	var response *meilisearch.SearchResponse
	err := s.withClient(func(client meilisearch.ServiceManager) error {
		var err error
		response, err = client.Index(indexUID).Search(query, request)
		return err
	})
	return response, err
}

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
