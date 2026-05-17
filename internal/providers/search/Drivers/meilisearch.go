package Drivers

import (
	"fmt"
	"net/http"
	"time"

	"fiber-starter/configs"
	"fiber-starter/internal/providers/search/Contracts"
	helpers "fiber-starter/internal/support"

	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchDriver struct {
	client meilisearch.ServiceManager
}

func NewMeilisearchDriver(cfg *configs.Config) *MeilisearchDriver {
	if cfg.Search.Host == "" {
		helpers.Warn("Meilisearch host is not configured, search operations will fail")
		return &MeilisearchDriver{}
	}

	client := meilisearch.New(
		cfg.Search.Host,
		meilisearch.WithAPIKey(cfg.Search.APIKey),
		meilisearch.WithCustomClient(&http.Client{Timeout: 5 * time.Second}),
	)

	return &MeilisearchDriver{client: client}
}

func (d *MeilisearchDriver) CreateIndex(uid string, primaryKey string) (*Contracts.TaskInfo, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.CreateIndex(&meilisearch.IndexConfig{Uid: uid, PrimaryKey: primaryKey})
	if err != nil {
		return nil, err
	}
	return &Contracts.TaskInfo{UID: int64(resp.TaskUID), Status: string(resp.Status)}, nil
}

func (d *MeilisearchDriver) DeleteIndex(uid string) (*Contracts.TaskInfo, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.DeleteIndex(uid)
	if err != nil {
		return nil, err
	}
	return &Contracts.TaskInfo{UID: int64(resp.TaskUID), Status: string(resp.Status)}, nil
}

func (d *MeilisearchDriver) AddDocuments(indexUID string, documents interface{}) (*Contracts.TaskInfo, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.Index(indexUID).AddDocuments(documents, nil)
	if err != nil {
		return nil, err
	}
	return &Contracts.TaskInfo{UID: int64(resp.TaskUID), Status: string(resp.Status)}, nil
}

func (d *MeilisearchDriver) UpdateDocuments(indexUID string, documents interface{}) (*Contracts.TaskInfo, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.Index(indexUID).UpdateDocuments(documents, nil)
	if err != nil {
		return nil, err
	}
	return &Contracts.TaskInfo{UID: int64(resp.TaskUID), Status: string(resp.Status)}, nil
}

func (d *MeilisearchDriver) DeleteDocuments(indexUID string, ids []string) (*Contracts.TaskInfo, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.Index(indexUID).DeleteDocuments(ids, nil)
	if err != nil {
		return nil, err
	}
	return &Contracts.TaskInfo{UID: int64(resp.TaskUID), Status: string(resp.Status)}, nil
}

func (d *MeilisearchDriver) DeleteAllDocuments(indexUID string) (*Contracts.TaskInfo, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.Index(indexUID).DeleteAllDocuments(nil)
	if err != nil {
		return nil, err
	}
	return &Contracts.TaskInfo{UID: int64(resp.TaskUID), Status: string(resp.Status)}, nil
}

func (d *MeilisearchDriver) Search(indexUID string, query string, request *Contracts.SearchRequest) (*Contracts.SearchResponse, error) {
	if d.client == nil {
		return nil, fmt.Errorf("meilisearch client not initialized")
	}
	if request == nil {
		request = &Contracts.SearchRequest{}
	}

	meiliReq := &meilisearch.SearchRequest{
		Offset:               request.Offset,
		Limit:                request.Limit,
		AttributesToRetrieve: request.AttributesToRetrieve,
	}

	if request.Filter != nil {
		meiliReq.Filter = request.Filter
	}
	if len(request.Sort) > 0 {
		meiliReq.Sort = request.Sort
	}

	resp, err := d.client.Index(indexUID).Search(query, meiliReq)
	if err != nil {
		return nil, err
	}

	hits := make([]interface{}, len(resp.Hits))
	for i, hit := range resp.Hits {
		hits[i] = hit
	}

	return &Contracts.SearchResponse{
		Hits:             hits,
		TotalHits:        resp.TotalHits,
		ProcessingTimeMs: resp.ProcessingTimeMs,
		Limit:            resp.Limit,
		Offset:           resp.Offset,
	}, nil
}

func (d *MeilisearchDriver) HealthCheck() error {
	if d.client == nil {
		return fmt.Errorf("meilisearch client not initialized")
	}
	resp, err := d.client.Health()
	if err != nil {
		return err
	}
	if resp.Status != "available" {
		return fmt.Errorf("meilisearch status: %s", resp.Status)
	}
	return nil
}

func (d *MeilisearchDriver) Close() error {
	return nil
}
