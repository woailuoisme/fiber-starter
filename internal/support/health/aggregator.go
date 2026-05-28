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
	Driver   string `json:"driver,omitempty"`
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

	cfg := a.app.AppConfig()

	getDriver := func(name string) string {
		if cfg == nil {
			return ""
		}
		switch name {
		case "database":
			defaultConn := cfg.Database.Default
			if conn, ok := cfg.Database.Connections[defaultConn]; ok {
				return conn.Driver
			}
			return defaultConn
		case "cache":
			return cfg.Cache.Driver
		case "mail":
			return cfg.Mail.Default
		case "queue":
			if cfg.Queue.Enabled {
				return "asynq"
			}
			return "noop"
		case "search":
			return cfg.Search.Default
		case "storage":
			return cfg.Storage.Driver
		default:
			return ""
		}
	}

	// List of providers to check
	checks := []struct {
		name     string
		enabled  bool
		critical bool
		driver   string
		check    func() error
	}{
		{
			name:     "database",
			enabled:  true,
			critical: a.isCritical("database", true),
			driver:   getDriver("database"),
			check: func() error {
				if a.app.ConnectionValue() == nil {
					return fmt.Errorf("database connection not initialized")
				}
				return a.app.ConnectionValue().HealthCheck()
			},
		},
		{
			name:     "cache",
			enabled:  cfg != nil && cfg.Cache.Enabled,
			critical: a.isCritical("cache", false),
			driver:   getDriver("cache"),
			check: func() error {
				if a.app.CacheStore() == nil {
					return fmt.Errorf("cache store not initialized")
				}
				return a.app.CacheStore().HealthCheck()
			},
		},
		{
			name:     "mail",
			enabled:  cfg != nil && cfg.Mail.Enabled,
			critical: a.isCritical("mail", false),
			driver:   getDriver("mail"),
			check: func() error {
				if a.app.EmailServiceValue() == nil {
					return fmt.Errorf("mail service not initialized")
				}
				return a.app.EmailServiceValue().HealthCheck()
			},
		},
		{
			name:     "queue",
			enabled:  cfg != nil && cfg.Queue.Enabled,
			critical: a.isCritical("queue", false),
			driver:   getDriver("queue"),
			check: func() error {
				if a.app.QueueServiceValue() == nil {
					return fmt.Errorf("queue service not initialized")
				}
				return a.app.QueueServiceValue().HealthCheck()
			},
		},
		{
			name:     "search",
			enabled:  cfg != nil && cfg.Search.Enabled,
			critical: a.isCritical("search", false),
			driver:   getDriver("search"),
			check: func() error {
				if a.app.SearchServiceValue() == nil {
					return fmt.Errorf("search service not initialized")
				}
				return a.app.SearchServiceValue().HealthCheck()
			},
		},
		{
			name:     "storage",
			enabled:  cfg != nil && cfg.Storage.Enabled,
			critical: a.isCritical("storage", false),
			driver:   getDriver("storage"),
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
				Driver:   c.driver,
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
				Driver:   c.driver,
				Error:    helpers.RedactSensitive(reason),
				Enabled:  true,
				Critical: c.critical,
			}
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(name string, critical bool, driver string, check func() error) {
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
				Driver:   driver,
				Error:    errMsg,
				Enabled:  true,
				Critical: critical,
			}
		}(c.name, c.critical, c.driver, c.check)
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
