package contracts

// SearchRequest represents a generic search request
type SearchRequest struct {
	Offset               int64
	Limit                int64
	Filter               interface{}
	Sort                 []string
	AttributesToRetrieve []string
}

// SearchResponse represents a generic search response
type SearchResponse struct {
	Hits             []interface{}
	TotalHits        int64
	ProcessingTimeMs int64
	Limit            int64
	Offset           int64
}

// TaskInfo represents a generic asynchronous task info
type TaskInfo struct {
	UID    int64
	Status string
}

// Engine defines the interface for search engines (similar to Laravel's Scout)
type Engine interface {
	// CreateIndex creates a new search index
	CreateIndex(uid string, primaryKey string) (*TaskInfo, error)

	// DeleteIndex deletes a search index
	DeleteIndex(uid string) (*TaskInfo, error)

	// AddDocuments adds documents to a search index
	AddDocuments(indexUID string, documents interface{}) (*TaskInfo, error)

	// UpdateDocuments updates documents in a search index
	UpdateDocuments(indexUID string, documents interface{}) (*TaskInfo, error)

	// DeleteDocuments deletes documents from a search index
	DeleteDocuments(indexUID string, ids []string) (*TaskInfo, error)

	// DeleteAllDocuments deletes all documents from a search index
	DeleteAllDocuments(indexUID string) (*TaskInfo, error)

	// Search performs a search query
	Search(indexUID string, query string, request *SearchRequest) (*SearchResponse, error)

	// HealthCheck checks the health of the search engine
	HealthCheck() error

	// Close closes the search engine connection
	Close() error
}
