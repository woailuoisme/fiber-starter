package search

import (
	"errors"
	"sync"

	"lfiber/configs"
	"lfiber/pkg/search/drive"
)

// ManagerImpl 实现了 Manager 接口，用于管理不同的搜索引擎实例的生命周期与缓存
type ManagerImpl struct {
	config  *configs.Config
	engines map[string]Engine
	mu      sync.Mutex
}

// NewManager 实例化一个搜索驱动管理器
func NewManager(cfg *configs.Config) *ManagerImpl {
	return &ManagerImpl{
		config:  cfg,
		engines: make(map[string]Engine),
	}
}

// Drive 返回指定的检索驱动实例（如 meilisearch、null）
func (m *ManagerImpl) Drive(name ...string) Engine {
	m.mu.Lock()
	defer m.mu.Unlock()

	driver := m.config.Search.Default
	if len(name) > 0 && name[0] != "" {
		driver = name[0]
	}

	if engine, ok := m.engines[driver]; ok {
		return engine
	}

	var engine Engine
	switch driver {
	case "meilisearch":
		engine = drive.NewMeilisearchDriver(m.config)
	case "null":
		engine = drive.NewNullDriver()
	default:
		engine = drive.NewMeilisearchDriver(m.config)
	}

	m.engines[driver] = engine
	return engine
}

// Close 关闭并释放所有已缓存的引擎连接
func (m *ManagerImpl) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for name, engine := range m.engines {
		if err := engine.Close(); err != nil {
			errs = append(errs, errors.New(name+": "+err.Error()))
		}
	}

	return errors.Join(errs...)
}
