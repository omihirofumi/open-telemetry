package open_telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"log"
)
import sdktrace "go.opentelemetry.io/otel/sdk/trace"

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
}
