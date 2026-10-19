package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_api_requests_total",
			Help: "Total HTTP requests by method, path, and status.",
		},
		[]string{"method", "path", "status"},
	)

	errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "product_api_errors_total",
			Help: "Total non-2xx HTTP responses.",
		},
		[]string{"method", "path"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "product_api_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(
		requestsTotal,
		errorsTotal,
		requestDuration,
	)
}

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		method := c.Request.Method

		requestsTotal.WithLabelValues(
			method, path, strconv.Itoa(status),
		).Inc()

		if status < http.StatusOK || status >= http.StatusMultipleChoices {
			errorsTotal.WithLabelValues(method, path).Inc()
		}

		requestDuration.WithLabelValues(method, path).Observe(
			time.Since(start).Seconds(),
		)
	}
}
