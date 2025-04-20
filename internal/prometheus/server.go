// Package prometheus это метрики приложения
package prometheus

import (
	"github.com/gofiber/fiber/v2"
	"internshipPVZ/internal/middleware"
)

// NewPrometheusServer конструктор
func NewPrometheusServer() *fiber.App {

	app := fiber.New()
	app.Get("/metrics", middleware.MetricsHandler())
	return app
}
