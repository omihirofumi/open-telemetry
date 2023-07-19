package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const serviceName = "anonymous"

func createAndRegisterExporters() error {
	stdExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
	)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(stdExporter),
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
