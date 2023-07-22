package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"

	api "go.opentelemetry.io/otel/metric"
)

const serviceName = "fibonacci"

var requests api.Int64Counter

var labels = api.WithAttributes(
	attribute.Key("application").String(serviceName),
	attribute.Key("container_id").String(os.Getenv("HOSTNAME")),
)

func main() {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	go serveMetrics()

	if err = buildRequestsCounter(); err != nil {
		log.Fatal(err)
	}

	if err = buildRuntimeObservers(); err != nil {
		log.Fatal(err)
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()
}

func serveMetrics() {
	log.Printf("serving metrics at localhost:2223/metrics")
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", http.HandlerFunc(fibHandler))

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Printf("error serving http: %v", err)
		return
	}
}

func buildRequestsCounter() error {
	var err error

	meter := otel.GetMeterProvider().Meter(serviceName)

	requests, err = meter.Int64Counter("fibonacci_requests_total",
		api.WithDescription("Total number of Fibonacci requests."))

	return err
}

func buildRuntimeObservers() error {
	var err error
	var m runtime.MemStats

	meter := otel.GetMeterProvider().Meter(serviceName)

	_, err = meter.Int64ObservableUpDownCounter("memory_usage_bytes",
		api.WithInt64Callback(func(_ context.Context, result api.Int64Observer) error {
			log.Println("memory_usage_bytes", int64(m.Sys))
			result.Observe(int64(m.Sys), labels)
			return nil
		}),
		api.WithDescription("Amount of memory used"),
		api.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableUpDownCounter("num_goroutines",
		api.WithInt64Callback(func(_ context.Context, result api.Int64Observer) error {
			log.Println("num_goroutines", int64(runtime.NumGoroutine()))
			result.Observe(int64(runtime.NumGoroutine()), labels)
			return nil
		}),
		api.WithDescription("Number of running goroutines."),
	)
	if err != nil {
		return err
	}

	return nil
}

func fibHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var n int

	if len(r.URL.Query()["n"]) != 1 {
		err = fmt.Errorf("wrong number of arguments")
	} else {
		n, err = strconv.Atoi(r.URL.Query()["n"][0])
	}

	if err != nil {
		http.Error(w, "couldn't parse index n", 400)
		return
	}

	ctx := r.Context()

	result := <-Fibonacci(ctx, n)

	fmt.Fprintln(w, result)
}

var requestCount int64

func Fibonacci(ctx context.Context, n int) chan int {
	requests.Add(ctx, 1, labels)

	atomic.AddInt64(&requestCount, 1)

	ch := make(chan int)

	go func() {
		result := 1
		if n > 1 {
			a := Fibonacci(ctx, n-1)
			b := Fibonacci(ctx, n-2)
			result = <-a + <-b
		}
		ch <- result
	}()

	return ch
}
