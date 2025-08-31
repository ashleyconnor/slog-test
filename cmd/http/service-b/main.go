package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"slog-test/pkg/otelinit"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

func main() {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	log.Println("Starting", serviceName)
	shutdown := otelinit.InitProvider(serviceName)
	defer shutdown(context.Background())

	tracer := otel.Tracer(serviceName)

	http.Handle("/world", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "handle-world")
		defer span.End()

		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.With("trace_id", span.SpanContext().TraceID().String())

		logger.Info("service-b span", "traceID", span.SpanContext().TraceID(), "spanID", span.SpanContext().SpanID())

		logger.Info("Received request from service-a")

		fmt.Fprintf(w, "Hello from Service: B")
	}), "world"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
