package search

import (
	"context"
	"sync"

	"github.com/uptrace/bun"
)

// SelectQueryFunc 产生对应模型查询的 SelectQuery 构造函数
type SelectQueryFunc func(ctx context.Context) (*bun.SelectQuery, error)

type registry struct {
	mu      sync.RWMutex
	models  map[string]Searchable
	queries map[string]SelectQueryFunc
}

var defaultRegistry = &registry{
	models:  make(map[string]Searchable),
	queries: make(map[string]SelectQueryFunc),
}

// RegisterModel 向搜索注册中心注册一个 Searchable 模型及其 SelectQuery 查询构造器
// 为什么这样做：在 Go 的静态类型中实现动态 CLI 导入（如 scout:import <model_name>）必须要由全局注册映射表提供支持
func RegisterModel(name string, model Searchable, queryFunc SelectQueryFunc) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.models[name] = model
	defaultRegistry.queries[name] = queryFunc
}

// Get 从注册中心获取指定名称的模型及其查询构造器
func Get(name string) (Searchable, SelectQueryFunc, bool) {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	model, ok1 := defaultRegistry.models[name]
	queryFunc, ok2 := defaultRegistry.queries[name]
	return model, queryFunc, ok1 && ok2
}

// Names 返回所有已注册的搜索模型的别名列表
func Names() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	names := make([]string, 0, len(defaultRegistry.models))
	for name := range defaultRegistry.models {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry 返回默认可用的搜索注册中心
func DefaultRegistry() *registry {
	return defaultRegistry
}
