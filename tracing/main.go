package main

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	jaegerEndpoint = "http://localhost:14268/api/traces"
	serviceName    = "fibonacci"
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
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
		),
	)

	otel.SetTracerProvider(tp)

	return nil
}

func main() {
	err := createAndRegisterExporters()
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", otelhttp.NewHandler(http.HandlerFunc(fibHandler), "root"))

	log.Fatal(http.ListenAndServe(":3000", nil))
}

func fibHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var n int

	if len(r.URL.Query()["n"]) != 1 {
		err = fmt.Errorf("wrong number of args")
	} else {
		n, err = strconv.Atoi(r.URL.Query()["n"][0])
	}

	if err != nil {
		http.Error(w, "couldn't parse index n", 400)
		return
	}

	ctx := r.Context()

	result := <-Fibonacci(ctx, n)

	sp := trace.SpanFromContext(ctx)
	sp.SetAttributes(attribute.Key("parameter").Int(n), attribute.Key("result").Int(result))

	fmt.Fprintln(w, result)
}

func Fibonacci(ctx context.Context, n int) chan int {
	ch := make(chan int)

	go func() {
		tr := otel.GetTracerProvider().Tracer(serviceName)

		cctx, sp := tr.Start(ctx,
			fmt.Sprintf("Fibonacci(%d)", n),
			trace.WithAttributes(attribute.Key("n").Int(n)))
		defer sp.End()

		result := 1
		if n > 1 {
			a := Fibonacci(cctx, n-1)
			b := Fibonacci(cctx, n-2)
			result = <-a + <-b
		}

		sp.SetAttributes(attribute.Key("result").Int(result))

		ch <- result
	}()
	return ch
}
