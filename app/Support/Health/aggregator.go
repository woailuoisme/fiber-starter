package health

import (
	"fmt"
	"sync"

	providers "fiber-starter/app/Providers"
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
	rt *providers.Runtime
}

// NewAggregator creates a new health aggregator
func NewAggregator(rt *providers.Runtime) *Aggregator {
	return &Aggregator{rt: rt}
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
				if a.rt.Connection == nil {
					return fmt.Errorf("database connection not initialized")
				}
				return a.rt.Connection.HealthCheck()
			},
		},
		{
			name:    "cache",
			enabled: a.rt.Config != nil && a.rt.Config.Cache.Enabled,
			check: func() error {
				if a.rt.Cache == nil {
					return fmt.Errorf("cache store not initialized")
				}
				return a.rt.Cache.HealthCheck()
			},
		},
		{
			name:    "mail",
			enabled: a.rt.Config != nil && a.rt.Config.Mail.Enabled,
			check: func() error {
				if a.rt.EmailService == nil {
					return fmt.Errorf("mail service not initialized")
				}
				return a.rt.EmailService.HealthCheck()
			},
		},
		{
			name:    "queue",
			enabled: a.rt.Config != nil && a.rt.Config.Queue.Enabled,
			check: func() error {
				if a.rt.QueueService == nil {
					return fmt.Errorf("queue service not initialized")
				}
				return a.rt.QueueService.HealthCheck()
			},
		},
		{
			name:    "search",
			enabled: a.rt.Config != nil && a.rt.Config.Search.Enabled,
			check: func() error {
				if a.rt.SearchService == nil {
					return fmt.Errorf("search service not initialized")
				}
				return a.rt.SearchService.HealthCheck()
			},
		},
		{
			name:    "storage",
			enabled: a.rt.Config != nil && a.rt.Config.Storage.Enabled,
			check: func() error {
				if a.rt.Storage == nil {
					return fmt.Errorf("storage manager not initialized")
				}
				return a.rt.Storage.Disk().HealthCheck()
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
