package health

import (
	"fmt"
	"sync"

	"fiber-starter/internal/support/appctx"
)

// Status represents the health status of a single provider
type Status struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Enabled bool   `json:"enabled"`
}

// Aggregator collects health status from all registered providers
type Aggregator struct {
	app appctx.Application
}

// NewAggregator creates a new health aggregator
func NewAggregator(app appctx.Application) *Aggregator {
	return &Aggregator{app: app}
}

// CheckAll returns the health status of all enabled providers
func (a *Aggregator) CheckAll() (map[string]Status, bool) {
	results := make(map[string]Status)
	allHealthy := true
	var mu sync.Mutex
	var wg sync.WaitGroup

	// List of providers to check
	checks := []struct {
		name    string
		enabled bool
		check   func() error
	}{
		{
			name:    "database",
			enabled: true,
			check: func() error {
				if a.app.ConnectionValue() == nil {
					return fmt.Errorf("database connection not initialized")
				}
				return a.app.ConnectionValue().HealthCheck()
			},
		},
		{
			name:    "cache",
			enabled: a.app.AppConfig() != nil && a.app.AppConfig().Cache.Enabled,
			check: func() error {
				if a.app.CacheStore() == nil {
					return fmt.Errorf("cache store not initialized")
				}
				return a.app.CacheStore().HealthCheck()
			},
		},
		{
			name:    "mail",
			enabled: a.app.AppConfig() != nil && a.app.AppConfig().Mail.Enabled,
			check: func() error {
				if a.app.EmailServiceValue() == nil {
					return fmt.Errorf("mail service not initialized")
				}
				return a.app.EmailServiceValue().HealthCheck()
			},
		},
		{
			name:    "queue",
			enabled: a.app.AppConfig() != nil && a.app.AppConfig().Queue.Enabled,
			check: func() error {
				if a.app.QueueServiceValue() == nil {
					return fmt.Errorf("queue service not initialized")
				}
				return a.app.QueueServiceValue().HealthCheck()
			},
		},
		{
			name:    "search",
			enabled: a.app.AppConfig() != nil && a.app.AppConfig().Search.Enabled,
			check: func() error {
				if a.app.SearchServiceValue() == nil {
					return fmt.Errorf("search service not initialized")
				}
				return a.app.SearchServiceValue().HealthCheck()
			},
		},
		{
			name:    "storage",
			enabled: a.app.AppConfig() != nil && a.app.AppConfig().Storage.Enabled,
			check: func() error {
				if a.app.StorageValue() == nil {
					return fmt.Errorf("storage manager not initialized")
				}
				return a.app.StorageValue().Disk().HealthCheck()
			},
		},
	}

	for _, c := range checks {
		if !c.enabled {
			results[c.name] = Status{
				Name:    c.name,
				Status:  "disabled",
				Enabled: false,
			}
			continue
		}

		wg.Add(1)
		go func(name string, check func() error) {
			defer wg.Done()
			err := safeCheck(check)

			mu.Lock()
			defer mu.Unlock()

			status := "ok"
			errMsg := ""
			if err != nil {
				status = "fail"
				errMsg = err.Error()
				allHealthy = false
			}

			results[name] = Status{
				Name:    name,
				Status:  status,
				Error:   errMsg,
				Enabled: true,
			}
		}(c.name, c.check)
	}

	wg.Wait()
	return results, allHealthy
}

func safeCheck(check func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("health check panic: %v", r)
		}
	}()
	return check()
}
