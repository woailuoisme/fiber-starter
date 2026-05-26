package health

import (
	"fmt"
	"sync"
	"time"

	"lfiber/configs"
	helpers "lfiber/internal/support"
	"lfiber/internal/support/appctx"
)

const (
	OverallOK       = "ok"
	OverallDegraded = "degraded"
	OverallFail     = "fail"

	StatusOK       = "ok"
	StatusDegraded = "degraded"
	StatusFail     = "fail"
	StatusDisabled = "disabled"

	defaultCheckTimeout = 2 * time.Second
)

// Status represents the health status of a single provider
type Status struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
	Enabled  bool   `json:"enabled"`
	Critical bool   `json:"critical"`
}

// Aggregator collects health status from all registered providers
type Aggregator struct {
	app appctx.Application
}

type degradedProviderReporter interface {
	DegradedProviders() map[string]string
}

// NewAggregator creates a new health aggregator
func NewAggregator(app appctx.Application) *Aggregator {
	return &Aggregator{app: app}
}

// CheckAll returns the health status of all enabled providers
func (a *Aggregator) CheckAll() (map[string]Status, bool) {
	results, overall := a.Check()
	return results, overall != OverallFail
}

// Check returns the health status of all known providers and an aggregate state.
func (a *Aggregator) Check() (map[string]Status, string) {
	results := make(map[string]Status)
	overall := OverallOK
	degradedProviders := a.degradedProviders()
	var mu sync.Mutex
	var wg sync.WaitGroup

	// List of providers to check
	checks := []struct {
		name     string
		enabled  bool
		critical bool
		check    func() error
	}{
		{
			name:     "database",
			enabled:  true,
			critical: a.isCritical("database", true),
			check: func() error {
				if a.app.ConnectionValue() == nil {
					return fmt.Errorf("database connection not initialized")
				}
				return a.app.ConnectionValue().HealthCheck()
			},
		},
		{
			name:     "cache",
			enabled:  a.app.AppConfig() != nil && a.app.AppConfig().Cache.Enabled,
			critical: a.isCritical("cache", false),
			check: func() error {
				if a.app.CacheStore() == nil {
					return fmt.Errorf("cache store not initialized")
				}
				return a.app.CacheStore().HealthCheck()
			},
		},
		{
			name:     "mail",
			enabled:  a.app.AppConfig() != nil && a.app.AppConfig().Mail.Enabled,
			critical: a.isCritical("mail", false),
			check: func() error {
				if a.app.EmailServiceValue() == nil {
					return fmt.Errorf("mail service not initialized")
				}
				return a.app.EmailServiceValue().HealthCheck()
			},
		},
		{
			name:     "queue",
			enabled:  a.app.AppConfig() != nil && a.app.AppConfig().Queue.Enabled,
			critical: a.isCritical("queue", false),
			check: func() error {
				if a.app.QueueServiceValue() == nil {
					return fmt.Errorf("queue service not initialized")
				}
				return a.app.QueueServiceValue().HealthCheck()
			},
		},
		{
			name:     "search",
			enabled:  a.app.AppConfig() != nil && a.app.AppConfig().Search.Enabled,
			critical: a.isCritical("search", false),
			check: func() error {
				if a.app.SearchServiceValue() == nil {
					return fmt.Errorf("search service not initialized")
				}
				return a.app.SearchServiceValue().HealthCheck()
			},
		},
		{
			name:     "storage",
			enabled:  a.app.AppConfig() != nil && a.app.AppConfig().Storage.Enabled,
			critical: a.isCritical("storage", false),
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
				Name:     c.name,
				Status:   StatusDisabled,
				Enabled:  false,
				Critical: c.critical,
			}
			continue
		}

		if reason, degraded := degradedProviders[c.name]; degraded {
			status := StatusDegraded
			mu.Lock()
			if c.critical {
				status = StatusFail
				overall = OverallFail
			} else if overall == OverallOK {
				overall = OverallDegraded
			}
			results[c.name] = Status{
				Name:     c.name,
				Status:   status,
				Error:    helpers.RedactSensitive(reason),
				Enabled:  true,
				Critical: c.critical,
			}
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(name string, critical bool, check func() error) {
			defer wg.Done()
			err := safeCheckWithTimeout(check, defaultCheckTimeout)

			mu.Lock()
			defer mu.Unlock()

			status := StatusOK
			errMsg := ""
			if err != nil {
				status = StatusDegraded
				if critical {
					status = StatusFail
					overall = OverallFail
				} else if overall == OverallOK {
					overall = OverallDegraded
				}
				errMsg = helpers.RedactError(err)
			}

			results[name] = Status{
				Name:     name,
				Status:   status,
				Error:    errMsg,
				Enabled:  true,
				Critical: critical,
			}
		}(c.name, c.critical, c.check)
	}

	wg.Wait()
	return results, overall
}

func (a *Aggregator) degradedProviders() map[string]string {
	reporter, ok := a.app.(degradedProviderReporter)
	if !ok {
		return nil
	}
	return reporter.DegradedProviders()
}

func (a *Aggregator) isCritical(name string, fallback bool) bool {
	cfg := a.app.AppConfig()
	if cfg == nil {
		return fallback
	}

	dependencies := cfg.Services.Dependencies
	if dependencies == nil {
		return fallback
	}

	dependency, ok := dependencies[name]
	if !ok {
		return fallback
	}
	return dependency.Critical
}

func safeCheckWithTimeout(check func() error, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- safeCheck(check)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("health check timed out after %s", timeout)
	}
}

func safeCheck(check func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("health check panic: %v", r)
		}
	}()
	return check()
}

func CriticalityFromConfig(cfg *configs.Config, name string, fallback bool) bool {
	if cfg == nil || cfg.Services.Dependencies == nil {
		return fallback
	}
	dependency, ok := cfg.Services.Dependencies[name]
	if !ok {
		return fallback
	}
	return dependency.Critical
}
