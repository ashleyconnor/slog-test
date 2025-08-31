package main

import (
	"context"
	"io"
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

	http.Handle("/hello", otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handle-hello")
		defer span.End()

		logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.With("trace_id", span.SpanContext().TraceID().String())

		logger.Info("service-a span", "traceID", span.SpanContext().TraceID(), "spanID", span.SpanContext().SpanID())

		// call service-b
		client := http.Client{
			Transport: otelhttp.NewTransport(http.DefaultTransport), // inject trace context
		}
		req, _ := http.NewRequestWithContext(ctx, "GET", "http://service-b:8080/world", nil)
		for key, value := range req.Header {
			log.Printf("service-a sending header %q: %q", key, value)
		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Error calling service-b", "err", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			logger.Info("response from service-b", "body", string(body))
		}

		logger.Info("Hello from Service: A")
	}), "hello"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
