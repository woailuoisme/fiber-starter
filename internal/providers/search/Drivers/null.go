package drivers

import (
	"fiber-starter/internal/providers/search/contracts"
)

type NullDriver struct{}

func NewNullDriver() *NullDriver {
	return &NullDriver{}
}

func (d *NullDriver) CreateIndex(uid string, primaryKey string) (*contracts.TaskInfo, error) {
	return &contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) GetIndex(uid string) (interface{}, error) {
	return nil, nil
}

func (d *NullDriver) DeleteIndex(uid string) (*contracts.TaskInfo, error) {
	return &contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) AddDocuments(indexUID string, documents interface{}) (*contracts.TaskInfo, error) {
	return &contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) UpdateDocuments(indexUID string, documents interface{}) (*contracts.TaskInfo, error) {
	return &contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) DeleteDocuments(indexUID string, ids []string) (*contracts.TaskInfo, error) {
	return &contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) DeleteAllDocuments(indexUID string) (*contracts.TaskInfo, error) {
	return &contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) Search(indexUID string, query string, request *contracts.SearchRequest) (*contracts.SearchResponse, error) {
	return &contracts.SearchResponse{}, nil
}

func (d *NullDriver) HealthCheck() error {
	return nil
}

func (d *NullDriver) Close() error {
	return nil
}
