package main

import (
	"context"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	jaegerEndpoint = "http://localhost:14268/api/traces"
	serviceName    = "anonymous"
)

func createAndRegisterExporters() error {
	stdExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return err
	}

	jaegerExporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
	)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(stdExporter),
		sdktrace.WithSyncer(jaegerExporter),
	)

	otel.SetTracerProvider(tp)

	return nil
}

func main() {
	err := createAndRegisterExporters()
	if err != nil {
		log.Fatal(err)
	}

	tr := otel.GetTracerProvider().Tracer(serviceName)

	ctx, sp := tr.Start(context.Background(), "main")
	defer sp.End()

	Foo(ctx)
}

func Foo(ctx context.Context) {
	tr := otel.GetTracerProvider().Tracer(serviceName)
	_, sp := tr.Start(ctx, "Foo")
	defer sp.End()
}
