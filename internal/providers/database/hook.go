package database

import (
	"context"
	"time"

	"lfiber/configs"
	helpers "lfiber/internal/support"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// QueryHook is a custom Bun query hook for measuring query duration,
// logging queries and marking slow queries in OTEL trace spans and application logs.
type QueryHook struct {
	config             *configs.Config
	connectionName     string
	logQueries         bool
	slowQueryThreshold time.Duration
}

// NewQueryHook creates a new QueryHook instance.
func NewQueryHook(cfg *configs.Config, connName string) *QueryHook {
	var logQueries bool
	slowThreshold := 200 * time.Millisecond // default threshold 200ms

	if cfg != nil {
		logQueries = cfg.Database.LogQueries
		if cfg.Database.SlowQueryThresholdMS > 0 {
			slowThreshold = time.Duration(cfg.Database.SlowQueryThresholdMS) * time.Millisecond
		}

		// Check connection-specific override
		if connCfg, ok := cfg.Database.Connections[connName]; ok {
			if connCfg.LogQueries != nil {
				logQueries = *connCfg.LogQueries
			}
			if connCfg.SlowQueryThresholdMS != nil && *connCfg.SlowQueryThresholdMS > 0 {
				slowThreshold = time.Duration(*connCfg.SlowQueryThresholdMS) * time.Millisecond
			}
		}
	}

	return &QueryHook{
		config:             cfg,
		connectionName:     connName,
		logQueries:         logQueries,
		slowQueryThreshold: slowThreshold,
	}
}

// BeforeQuery is called before query execution.
func (h *QueryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

// AfterQuery is called after query execution.
func (h *QueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	duration := time.Since(event.StartTime)
	isSlow := duration >= h.slowQueryThreshold

	// 1. Log query if required or if it is slow
	if isSlow {
		helpers.Warn(
			"Slow query detected",
			zap.String("connection", h.connectionName),
			zap.String("query", event.Query),
			zap.Duration("duration", duration),
			zap.String("threshold", h.slowQueryThreshold.String()),
		)
	} else if h.logQueries {
		helpers.Info(
			"Query executed",
			zap.String("connection", h.connectionName),
			zap.String("query", event.Query),
			zap.Duration("duration", duration),
		)
	}

	// 2. Enrich OTEL span if trace is enabled
	if h.config != nil && h.config.OTEL.TraceEnabled {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			if isSlow {
				span.SetAttributes(
					attribute.Bool("db.slow_query", true),
					attribute.Int64("db.query_duration_ms", duration.Milliseconds()),
				)
				span.SetStatus(codes.Error, "slow query detected")
				span.AddEvent("slow_query", trace.WithAttributes(
					attribute.String("db.statement", event.Query),
					attribute.Int64("duration_ms", duration.Milliseconds()),
				))
			}
		}
	}
}

// NewQueryHookForTest_LogQueries is a test helper to inspect private field.
func (h *QueryHook) NewQueryHookForTest_LogQueries() bool {
	return h.logQueries
}

// NewQueryHookForTest_SlowQueryThreshold is a test helper to inspect private field.
func (h *QueryHook) NewQueryHookForTest_SlowQueryThreshold() time.Duration {
	return h.slowQueryThreshold
}
