package Drivers

import (
	"fiber-starter/app/Providers/Search/Contracts"
)

type NullDriver struct{}

func NewNullDriver() *NullDriver {
	return &NullDriver{}
}

func (d *NullDriver) CreateIndex(uid string, primaryKey string) (*Contracts.TaskInfo, error) {
	return &Contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) GetIndex(uid string) (interface{}, error) {
	return nil, nil
}

func (d *NullDriver) DeleteIndex(uid string) (*Contracts.TaskInfo, error) {
	return &Contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) AddDocuments(indexUID string, documents interface{}) (*Contracts.TaskInfo, error) {
	return &Contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) UpdateDocuments(indexUID string, documents interface{}) (*Contracts.TaskInfo, error) {
	return &Contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) DeleteDocuments(indexUID string, ids []string) (*Contracts.TaskInfo, error) {
	return &Contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) DeleteAllDocuments(indexUID string) (*Contracts.TaskInfo, error) {
	return &Contracts.TaskInfo{Status: "enqueued"}, nil
}

func (d *NullDriver) Search(indexUID string, query string, request *Contracts.SearchRequest) (*Contracts.SearchResponse, error) {
	return &Contracts.SearchResponse{}, nil
}

func (d *NullDriver) HealthCheck() error {
	return nil
}

func (d *NullDriver) Close() error {
	return nil
}
