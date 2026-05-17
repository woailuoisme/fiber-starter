package search

import (
	"errors"

	"fiber-starter/internal/providers/search/contracts"
	"fiber-starter/internal/support/appctx"
)

var ErrContainerNotInitialized = errors.New("application container not initialized")

// engine returns the default search engine instance from the container.
func engine() contracts.Engine {
	if app := appctx.App(); app != nil {
		return app.SearchServiceValue()
	}
	return nil
}

// manager returns the search manager instance from the container.
func manager() contracts.Manager {
	if app := appctx.App(); app != nil {
		return app.SearchManagerValue()
	}
	return nil
}

// Drive returns a specific search engine instance
func Drive(name ...string) contracts.Engine {
	if m := manager(); m != nil {
		return m.Drive(name...)
	}
	return nil
}

// CreateIndex creates a new search index using the default engine
func CreateIndex(uid string, primaryKey string) (*contracts.TaskInfo, error) {
	if e := engine(); e != nil {
		return e.CreateIndex(uid, primaryKey)
	}
	return nil, ErrContainerNotInitialized
}

// DeleteIndex deletes a search index using the default engine
func DeleteIndex(uid string) (*contracts.TaskInfo, error) {
	if e := engine(); e != nil {
		return e.DeleteIndex(uid)
	}
	return nil, ErrContainerNotInitialized
}

// Search performs a search query using the default engine
func Search(indexUID string, query string, request *contracts.SearchRequest) (*contracts.SearchResponse, error) {
	if e := engine(); e != nil {
		return e.Search(indexUID, query, request)
	}
	return nil, ErrContainerNotInitialized
}

// AddDocuments adds documents to a search index using the default engine
func AddDocuments(indexUID string, documents interface{}) (*contracts.TaskInfo, error) {
	if e := engine(); e != nil {
		return e.AddDocuments(indexUID, documents)
	}
	return nil, ErrContainerNotInitialized
}

// UpdateDocuments updates documents in a search index using the default engine
func UpdateDocuments(indexUID string, documents interface{}) (*contracts.TaskInfo, error) {
	if e := engine(); e != nil {
		return e.UpdateDocuments(indexUID, documents)
	}
	return nil, ErrContainerNotInitialized
}

// DeleteDocuments deletes documents from a search index using the default engine
func DeleteDocuments(indexUID string, ids []string) (*contracts.TaskInfo, error) {
	if e := engine(); e != nil {
		return e.DeleteDocuments(indexUID, ids)
	}
	return nil, ErrContainerNotInitialized
}

// DeleteAllDocuments deletes all documents from a search index using the default engine
func DeleteAllDocuments(indexUID string) (*contracts.TaskInfo, error) {
	if e := engine(); e != nil {
		return e.DeleteAllDocuments(indexUID)
	}
	return nil, ErrContainerNotInitialized
}

// HealthCheck checks the health of the default search engine
func HealthCheck() error {
	if e := engine(); e != nil {
		return e.HealthCheck()
	}
	return ErrContainerNotInitialized
}

// Close closes the default search engine connection
func Close() error {
	if m := manager(); m != nil {
		return m.Close()
	}
	return ErrContainerNotInitialized
}
