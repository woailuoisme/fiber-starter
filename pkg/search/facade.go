package search

import (
	"context"
	"errors"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

var (
	defaultManager Manager
	defaultEngine  Engine
)

// SetDefaultManager sets the global default SearchManager instance.
func SetDefaultManager(m Manager) {
	defaultManager = m
}

// SetDefaultEngine sets the global default Engine instance.
func SetDefaultEngine(e Engine) {
	defaultEngine = e
}

// engine returns the underlying cached default search engine instance
func engine() Engine {
	return defaultEngine
}

// manager returns the underlying SearchManager instance
func manager() Manager {
	return defaultManager
}

// Drive 获取特定指定的搜索引擎驱动
func Drive(name ...string) Engine {
	if m := manager(); m != nil {
		return m.Drive(name...)
	}
	return nil
}

// Index 将一个实现了 Searchable 契约的对象同步/索引到默认搜索引擎中（Scout 核心）
// 为什么这样做：完全屏蔽底层的 Meilisearch API 细节，让业务层直接声明式归档模型
func Index(ctx context.Context, model Searchable) (*TaskInfo, error) {
	e := engine()
	if e == nil {
		return nil, ErrContainerNotInitialized
	}

	indexUID := model.SearchableIndex()
	doc := model.ToSearchableArray()

	// 确保主键 ID 存入文档，以防止 Meilisearch 无法匹配唯一标识
	if _, exists := doc["id"]; !exists {
		doc["id"] = model.SearchableId()
	}

	return e.AddDocuments(indexUID, []any{doc})
}

// Delete 从搜索引擎的索引中移除指定模型的全文数据
func Delete(ctx context.Context, model Searchable) (*TaskInfo, error) {
	e := engine()
	if e == nil {
		return nil, ErrContainerNotInitialized
	}

	indexUID := model.SearchableIndex()
	return e.DeleteDocuments(indexUID, []string{model.SearchableId()})
}

// CreateIndex 创建搜索引擎索引
func CreateIndex(uid string, primaryKey string) (*TaskInfo, error) {
	if e := engine(); e != nil {
		return e.CreateIndex(uid, primaryKey)
	}
	return nil, ErrContainerNotInitialized
}

// DeleteIndex 彻底删除检索索引
func DeleteIndex(uid string) (*TaskInfo, error) {
	if e := engine(); e != nil {
		return e.DeleteIndex(uid)
	}
	return nil, ErrContainerNotInitialized
}

// Search 执行全文搜索查询
func Search(indexUID string, query string, request *SearchRequest) (*SearchResponse, error) {
	if e := engine(); e != nil {
		return e.Search(indexUID, query, request)
	}
	return nil, ErrContainerNotInitialized
}

// AddDocuments 直接推送原始文档数据
func AddDocuments(indexUID string, documents any) (*TaskInfo, error) {
	if e := engine(); e != nil {
		return e.AddDocuments(indexUID, documents)
	}
	return nil, ErrContainerNotInitialized
}

// UpdateDocuments 增量更新原始文档数据
func UpdateDocuments(indexUID string, documents any) (*TaskInfo, error) {
	if e := engine(); e != nil {
		return e.UpdateDocuments(indexUID, documents)
	}
	return nil, ErrContainerNotInitialized
}

// DeleteDocuments 批量删除原始文档
func DeleteDocuments(indexUID string, ids []string) (*TaskInfo, error) {
	if e := engine(); e != nil {
		return e.DeleteDocuments(indexUID, ids)
	}
	return nil, ErrContainerNotInitialized
}

// DeleteAllDocuments 清空指定索引的全部文档
func DeleteAllDocuments(indexUID string) (*TaskInfo, error) {
	if e := engine(); e != nil {
		return e.DeleteAllDocuments(indexUID)
	}
	return nil, ErrContainerNotInitialized
}

// HealthCheck 检查连通性
func HealthCheck() error {
	if e := engine(); e != nil {
		return e.HealthCheck()
	}
	return ErrContainerNotInitialized
}

// Close 释放连接
func Close() error {
	if m := manager(); m != nil {
		return m.Close()
	}
	return ErrContainerNotInitialized
}
