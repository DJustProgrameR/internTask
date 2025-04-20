// Package http это http сервер
package http

import (
	"github.com/gofiber/fiber/v2"
	"internshipPVZ/internal/http/handlers"
	"internshipPVZ/internal/middleware"
	"log"
)

// AuthController -
type AuthController interface {
	DummyLogin(ctx *fiber.Ctx) error

	Register(ctx *fiber.Ctx) error

	Login(ctx *fiber.Ctx) error
}

// ProductController -
type ProductController interface {
	CreateProduct(ctx *fiber.Ctx) error
}

// PVZController -
type PVZController interface {
	CreatePVZ(ctx *fiber.Ctx) error
	DeleteLastProduct(ctx *fiber.Ctx) error
	GetPVZs(ctx *fiber.Ctx) error
	CloseLastReception(ctx *fiber.Ctx) error
}

// ReceptionController -
type ReceptionController interface {
	CreateReception(ctx *fiber.Ctx) error
}

// JWTService токены
type JWTService interface {
	GetClaims(tokenString string) (string, error)
}

// NewHTTPServer конструктор
func NewHTTPServer(
	authController AuthController,
	pvzController PVZController,
	receptionController ReceptionController,
	productController ProductController,
	jwtService JWTService,
	logger handlers.Logger,
) *fiber.App {
	if authController == nil {
		log.Fatalf("HttpServer initialization failed: authController is nil")
	}
	if pvzController == nil {
		log.Fatalf("HttpServer initialization failed: pvzController is nil")
	}
	if receptionController == nil {
		log.Fatalf("HttpServer initialization failed: receptionController is nil")
	}
	if productController == nil {
		log.Fatalf("HttpServer initialization failed: productController is nil")
	}
	if jwtService == nil {
		log.Fatalf("HttpServer initialization failed: jwtService is nil")
	}
	if logger == nil {
		log.Fatalf("HttpServer initialization failed: logger is nil")
	}

	app := fiber.New()

	app.Use(middleware.AuthMiddleware(jwtService, logger))
	app.Use(middleware.PrometheusMiddleware(logger))

	app.Post("/dummyLogin", authController.DummyLogin)
	app.Post("/register", authController.Register)
	app.Post("/login", authController.Login)
	app.Post("/pvz", pvzController.CreatePVZ)
	app.Get("/pvz", pvzController.GetPVZs)
	app.Post("/pvz/:pvzId/close_last_reception", pvzController.CloseLastReception)
	app.Post("/pvz/:pvzId/delete_last_product", pvzController.DeleteLastProduct)
	app.Post("/receptions", receptionController.CreateReception)
	app.Post("/products", productController.CreateProduct)

	return app
}
