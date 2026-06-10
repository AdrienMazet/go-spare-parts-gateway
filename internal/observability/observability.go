package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spare_parts_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "spare_parts_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
	externalProviderRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spare_parts_external_provider_requests_total",
			Help: "Total number of external provider HTTP requests.",
		},
		[]string{"provider", "status"},
	)
	externalProviderRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "spare_parts_external_provider_request_duration_seconds",
			Help:    "External provider HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "status"},
	)
	dbOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spare_parts_db_operations_total",
			Help: "Total number of database operations.",
		},
		[]string{"operation", "status"},
	)
	dbOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "spare_parts_db_operation_duration_seconds",
			Help:    "Database operation duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "status"},
	)
	kafkaEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spare_parts_kafka_events_total",
			Help: "Total number of Kafka events handled by this service.",
		},
		[]string{"operation", "topic", "status"},
	)
)

// ConfigureLogger configures slog for structured JSON logs.
func ConfigureLogger(logLevel string) {
	level := slog.LevelInfo
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})))
}

// InitTracer configures OpenTelemetry trace export.
func InitTracer(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider.Shutdown, nil
}

// Tracer returns the application tracer.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// MetricsHandler returns the Prometheus scrape handler.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// HTTPMiddleware records HTTP request metrics.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		status := strconv.Itoa(recorder.status)
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(time.Since(startedAt).Seconds())
	})
}

// InstrumentHTTPHandler instruments an HTTP handler with OpenTelemetry.
func InstrumentHTTPHandler(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

// InstrumentHTTPClient returns an HTTP client transport instrumented with OpenTelemetry.
func InstrumentHTTPClientTransport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}

	return otelhttp.NewTransport(base)
}

// RecordExternalProviderRequest records external provider request metrics.
func RecordExternalProviderRequest(provider, status string, duration time.Duration) {
	externalProviderRequestsTotal.WithLabelValues(provider, status).Inc()
	externalProviderRequestDuration.WithLabelValues(provider, status).Observe(duration.Seconds())
}

// RecordDBOperation records database operation metrics.
func RecordDBOperation(operation string, err error, duration time.Duration) {
	status := "success"
	if err != nil {
		status = "error"
	}

	dbOperationsTotal.WithLabelValues(operation, status).Inc()
	dbOperationDuration.WithLabelValues(operation, status).Observe(duration.Seconds())
}

// RecordKafkaEvent records Kafka event metrics.
func RecordKafkaEvent(operation, topic string, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	kafkaEventsTotal.WithLabelValues(operation, topic, status).Inc()
}

// TraceAttributes builds common span attributes.
func TraceAttributes(values ...attribute.KeyValue) []attribute.KeyValue {
	return values
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}

	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}
