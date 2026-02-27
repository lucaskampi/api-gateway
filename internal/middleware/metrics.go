package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

func Metrics() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Path() == "/metrics" {
			return c.Next()
		}

		httpRequestsInFlight.Inc()
		start := time.Now()
		defer func() {
			httpRequestsInFlight.Dec()
			duration := time.Since(start).Seconds()
			method := c.Method()
			path := c.Route().Path
			status := c.Response().StatusCode()

			httpRequestsTotal.WithLabelValues(method, path, statusToString(status)).Inc()
			httpRequestDuration.WithLabelValues(method, path).Observe(duration)
		}()

		return c.Next()
	}
}

func statusToString(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	default:
		return "other"
	}
}
