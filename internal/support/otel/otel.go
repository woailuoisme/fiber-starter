package otel

import (
	"context"
	"fmt"

	"lfiber/configs"
	helpers "lfiber/internal/support"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.uber.org/zap"
)

// GlobalPrometheusExporter is the global exporter instance used to serve the /metrics endpoint
var GlobalPrometheusExporter *prometheus.Exporter

// InitOTEL 初始化 OpenTelemetry 追踪器和/或指标度量器。
// 返回一个关闭函数，应在应用程序退出时调用。
func InitOTEL(cfg *configs.Config) (func(context.Context) error, error) {
	if !cfg.OTEL.TraceEnabled && !cfg.OTEL.MetricsEnabled {
		return func(context.Context) error { return nil }, nil
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.OTEL.ServiceName),
	)

	var tp *sdktrace.TracerProvider
	var traceExporter sdktrace.SpanExporter
	var err error

	if cfg.OTEL.TraceEnabled {
		switch cfg.OTEL.ExporterType {
		case "otlp":
			traceExporter, err = otlptracegrpc.New(
				context.Background(),
				otlptracegrpc.WithEndpoint(cfg.OTEL.Endpoint),
				otlptracegrpc.WithInsecure(),
			)
		case "stdout":
			fallthrough
		default:
			traceExporter, err = stdout.New(stdout.WithPrettyPrint())
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create otel trace exporter: %w", err)
		}

		// Initialize TracerProvider
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithResource(res),
		)

		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	var mp *metric.MeterProvider
	if cfg.OTEL.MetricsEnabled {
		// Initialize Prometheus metric exporter
		metricExporter, err := prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create otel prometheus exporter: %w", err)
		}
		GlobalPrometheusExporter = metricExporter

		// Register exporter with MeterProvider
		mp = metric.NewMeterProvider(
			metric.WithReader(metricExporter),
			metric.WithResource(res),
		)
		otel.SetMeterProvider(mp)
	}

	helpers.Info(
		"otel_initialized",
		zap.String("exporter_type", cfg.OTEL.ExporterType),
		zap.Bool("trace_enabled", cfg.OTEL.TraceEnabled),
		zap.Bool("metrics_enabled", cfg.OTEL.MetricsEnabled),
	)

	return func(ctx context.Context) error {
		var shutdownErr error
		if tp != nil {
			if err := tp.Shutdown(ctx); err != nil {
				shutdownErr = fmt.Errorf("failed to shutdown tracer provider: %w", err)
			}
		}
		if mp != nil {
			if err := mp.Shutdown(ctx); err != nil {
				if shutdownErr != nil {
					shutdownErr = fmt.Errorf("%v; failed to shutdown meter provider: %w", shutdownErr, err)
				} else {
					shutdownErr = fmt.Errorf("failed to shutdown meter provider: %w", err)
				}
			}
		}
		return shutdownErr
	}, nil
}
