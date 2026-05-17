package otel

import (
	"context"
	"fmt"

	"fiber-starter/configs"
	helpers "fiber-starter/internal/support"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.uber.org/zap"
)

// InitOTEL 初始化 OpenTelemetry 追踪器。
// 返回一个关闭函数，应在应用程序退出时调用。
func InitOTEL(cfg *configs.Config) (func(context.Context) error, error) {
	if !cfg.OTEL.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	var exporter sdktrace.SpanExporter
	var err error

	switch cfg.OTEL.ExporterType {
	case "otlp":
		exporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(cfg.OTEL.Endpoint),
			otlptracegrpc.WithInsecure(),
		)
	case "stdout":
		fallthrough
	default:
		exporter, err = stdout.New(stdout.WithPrettyPrint())
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create otel exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.OTEL.ServiceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	helpers.Info("otel_initialized", zap.String("exporter_type", cfg.OTEL.ExporterType))

	return tp.Shutdown, nil
}
