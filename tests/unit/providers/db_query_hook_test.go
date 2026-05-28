package providers_test

import (
	"context"
	"testing"
	"time"

	"lfiber/configs"
	database "lfiber/internal/providers/database"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// mockSpan implements the trace.Span interface for testing purposes.
type mockSpan struct {
	trace.Span
	attributes []attribute.KeyValue
	status     codes.Code
	statusDesc string
	events     []string
}

func (m *mockSpan) IsRecording() bool {
	return true
}

func (m *mockSpan) SetAttributes(kvs ...attribute.KeyValue) {
	m.attributes = append(m.attributes, kvs...)
}

func (m *mockSpan) SetStatus(code codes.Code, desc string) {
	m.status = code
	m.statusDesc = desc
}

func (m *mockSpan) AddEvent(name string, _ ...trace.EventOption) {
	m.events = append(m.events, name)
}

func TestQueryHook_ConfigParsing(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Database.LogQueries = true
	cfg.Database.SlowQueryThresholdMS = 150

	// 1. Test global configs
	hook := database.NewQueryHook(cfg, "default")
	assert.True(t, hook.NewQueryHookForTest_LogQueries())
	assert.Equal(t, 150*time.Millisecond, hook.NewQueryHookForTest_SlowQueryThreshold())

	// 2. Test connection-specific override
	cfg.Database.Connections = map[string]configs.DBConnection{
		"custom": {
			LogQueries:           boolPtr(false),
			SlowQueryThresholdMS: intPtr(50),
		},
	}
	hookCustom := database.NewQueryHook(cfg, "custom")
	assert.False(t, hookCustom.NewQueryHookForTest_LogQueries())
	assert.Equal(t, 50*time.Millisecond, hookCustom.NewQueryHookForTest_SlowQueryThreshold())
}

func TestQueryHook_SlowQueryDetection(t *testing.T) {
	cfg := &configs.Config{}
	cfg.Database.LogQueries = false
	cfg.Database.SlowQueryThresholdMS = 50 // 50ms threshold
	cfg.OTEL.TraceEnabled = true

	hook := database.NewQueryHook(cfg, "default")

	// Create context with mock span
	mSpan := &mockSpan{}
	ctx := trace.ContextWithSpan(context.Background(), mSpan)

	// Execute event: slow query
	now := time.Now()
	eventSlow := &bun.QueryEvent{
		StartTime: now.Add(-100 * time.Millisecond), // 100ms duration
		Query:     "SELECT * FROM users",
	}
	hook.AfterQuery(ctx, eventSlow)

	// Verify span was marked as slow/error
	assert.Equal(t, codes.Error, mSpan.status)
	assert.Equal(t, "slow query detected", mSpan.statusDesc)
	assert.Contains(t, mSpan.events, "slow_query")

	hasSlowAttr := false
	for _, attr := range mSpan.attributes {
		if attr.Key == "db.slow_query" && attr.Value.AsBool() {
			hasSlowAttr = true
		}
	}
	assert.True(t, hasSlowAttr)

	// Execute event: fast query
	mSpanFast := &mockSpan{}
	ctxFast := trace.ContextWithSpan(context.Background(), mSpanFast)
	eventFast := &bun.QueryEvent{
		StartTime: now.Add(-10 * time.Millisecond), // 10ms duration
		Query:     "SELECT * FROM users LIMIT 1",
	}
	hook.AfterQuery(ctxFast, eventFast)

	// Verify fast query not marked
	assert.Equal(t, codes.Unset, mSpanFast.status)
	assert.Empty(t, mSpanFast.attributes)
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
