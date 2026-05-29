package contracts

// Searchable 允许一个 struct 在搜索引擎中执行全文检索同步（类似于 Laravel Scout）
type Searchable interface {
	// SearchableId 返回本记录在搜索引擎中的唯一 ID 标识（通常为主键）
	SearchableId() string
	// SearchableIndex 返回该模型对应的搜索引擎 Index 索引名称
	SearchableIndex() string
	// ToSearchableArray 将该记录转换为需要同步进搜索引擎的属性 map 集合
	ToSearchableArray() map[string]any
}

// SearchRequest 代表通用的全文检索请求参数
type SearchRequest struct {
	Offset               int64
	Limit                int64
	Filter               any
	Sort                 []string
	AttributesToRetrieve []string
}

// SearchResponse 代表通用的全文检索结果响应
type SearchResponse struct {
	Hits             []any
	TotalHits        int64
	ProcessingTimeMs int64
	Limit            int64
	Offset           int64
}

// TaskInfo 代表搜索引擎异步提交任务状态元数据
type TaskInfo struct {
	UID    int64
	Status string
}

// Engine 定义全文搜索引擎的基本 CRUD 接口（类似于 Laravel Scout 驱动）
type Engine interface {
	// CreateIndex 创建一个新的检索索引
	CreateIndex(uid string, primaryKey string) (*TaskInfo, error)

	// DeleteIndex 彻底删除一个索引
	DeleteIndex(uid string) (*TaskInfo, error)

	// AddDocuments 向指定的索引中追加/同步数据文档
	AddDocuments(indexUID string, documents any) (*TaskInfo, error)

	// UpdateDocuments 增量更新指定索引中的文档
	UpdateDocuments(indexUID string, documents any) (*TaskInfo, error)

	// DeleteDocuments 从指定的索引中删除指定的文档列表
	DeleteDocuments(indexUID string, ids []string) (*TaskInfo, error)

	// DeleteAllDocuments 清空指定索引下的全部文档
	DeleteAllDocuments(indexUID string) (*TaskInfo, error)

	// Search 执行全文搜索查询
	Search(indexUID string, query string, request *SearchRequest) (*SearchResponse, error)

	// HealthCheck 检查搜索引擎连通状态
	HealthCheck() error

	// Close 关闭并释放搜索引擎链接连接
	Close() error
}

// Manager 定义搜索多驱动管理接口
type Manager interface {
	// Drive 返回指定的全文检索驱动实例
	Drive(name ...string) Engine
	// Close 关闭所有已缓存的驱动实例
	Close() error
}
