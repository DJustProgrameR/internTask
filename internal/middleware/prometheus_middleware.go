// Package middleware это мидлвэр для приложения
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/http/onlymodels"
	"time"
)

var (
	// HTTPRequestsTotal общее кол-во запросов
	HTTPRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
	)

	// HTTPResponseTime время ответа
	HTTPResponseTime = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_response_time_milliseconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{5, 10, 25, 50, 75, 100, 200, 500},
		},
	)

	// PVZCreatedCount -
	PVZCreatedCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of PVZ created",
		},
	)

	// ReceptionsCreatedCount -
	ReceptionsCreatedCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total number of order receptions created",
		},
	)

	// ProductsAddedCount -
	ProductsAddedCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of products added",
		},
	)
)

func init() {
	prometheus.MustRegister(HTTPResponseTime)
	prometheus.MustRegister(PVZCreatedCount)
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(ProductsAddedCount)
	prometheus.MustRegister(ReceptionsCreatedCount)
}

// PrometheusMiddleware возвращает хэндлер для сбора метрик
func PrometheusMiddleware(logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c == nil {
			logger.Error("received nil ctx",
				"middleware", "PrometheusMiddleware",
				"error")
			return c.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
		}
		start := time.Now()
		path := c.Path()

		err := c.Next()

		status := c.Response().StatusCode()
		method := c.Method()

		HTTPRequestsTotal.Inc()
		HTTPResponseTime.Observe(float64(time.Since(start).Milliseconds()))
		if method == "POST" && path == "/pvz" && status == 201 {
			PVZCreatedCount.Inc()
		}
		if method == "POST" && path == "/receptions" && status == 201 {
			ReceptionsCreatedCount.Inc()
		}
		if method == "POST" && path == "/products" && status == 201 {
			ProductsAddedCount.Inc()
		}
		return err
	}
}

// MetricsHandler возвращает хэндлер для выдачи метрик
func MetricsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		handler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
		handler(c.Context())
		return nil
	}
}
