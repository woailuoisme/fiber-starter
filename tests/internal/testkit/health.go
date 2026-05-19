package testkit

import (
	"database/sql"

	"github.com/uptrace/bun"
)

// HealthCheckFunc adapts a function into a reusable health-check stub.
type HealthCheckFunc func() error

func (f HealthCheckFunc) HealthCheck() error {
	return f()
}

// StubConnection is a small database connection test double for readiness tests.
type StubConnection struct {
	Name        string
	DriverName  string
	DialectName string
	DB          *sql.DB
	Bun         *bun.DB
	HealthErr   error
	Stats       map[string]interface{}
	CloseErr    error
}

func (s *StubConnection) GetDB() (*sql.DB, error) {
	return s.DB, nil
}

func (s *StubConnection) BunDB() (*bun.DB, error) {
	return s.Bun, nil
}

func (s *StubConnection) Dialect() (string, error) {
	if s.DialectName == "" {
		return "stub", nil
	}
	return s.DialectName, nil
}

func (s *StubConnection) GetName() string {
	if s.Name == "" {
		return "stub"
	}
	return s.Name
}

func (s *StubConnection) GetDriverName() string {
	if s.DriverName == "" {
		return "stub"
	}
	return s.DriverName
}

func (s *StubConnection) HealthCheck() error {
	return s.HealthErr
}

func (s *StubConnection) GetStats() (map[string]interface{}, error) {
	if s.Stats == nil {
		return map[string]interface{}{}, nil
	}
	return s.Stats, nil
}

func (s *StubConnection) Close() error {
	return s.CloseErr
}
