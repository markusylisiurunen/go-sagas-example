package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

func newJaegerExporter(ctx context.Context) (*otlptrace.Exporter, error) {
	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	)
	return otlptrace.New(ctx, client)
}

func newTracerResource() (*resource.Resource, error) {
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("sagas"),
			semconv.ServiceVersionKey.String("v0.1.0"),
			semconv.ServiceInstanceIDKey.String(uuid.NewString()),
		),
	)
}

func setupTelemetry(ctx context.Context) error {
	exporter, err := newJaegerExporter(ctx)
	if err != nil {
		return err
	}
	res, err := newTracerResource()
	if err != nil {
		return err
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case <-sig:
		}
		if err := tp.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return nil
}
