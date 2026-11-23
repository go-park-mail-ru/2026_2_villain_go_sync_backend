package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	registry *prometheus.Registry
}

func NewHTTPMetrics() *HTTPMetrics {
	registry := prometheus.NewRegistry()
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_server_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:                        "http_server_request_duration_seconds",
			Help:                        "HTTP request duration in seconds",
			NativeHistogramBucketFactor: 1.1,
		},
		[]string{"method", "route", "status"},
	)

	registry.MustRegister(
		requests,
		duration,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &HTTPMetrics{
		requests: requests,
		duration: duration,
		registry: registry,
	}
}

func (m *HTTPMetrics) Observe(method string, route string, status int, duration time.Duration) {
	statusCode := strconv.Itoa(status)
	m.requests.WithLabelValues(method, route, statusCode).Inc()
	m.duration.WithLabelValues(method, route, statusCode).Observe(duration.Seconds())
}

func (m *HTTPMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
