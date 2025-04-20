// Package handlers это http хэндлеры
package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/http/onlymodels"
	"log"
)

// AuthUseCase интерфейс для аутентификации
type AuthUseCase interface {
	DummyLogin(ctx context.Context, request *onlymodels.PostDummyLoginJSONBody) (string, error)
	Register(ctx context.Context, request *onlymodels.PostRegisterJSONBody) (*onlymodels.User, error)
	Login(ctx context.Context, request *onlymodels.PostLoginJSONBody) (string, error)
}

// AuthController контроллер для аутентификации
type AuthController struct {
	authUsecase AuthUseCase
	logger      Logger
}

// NewAuthController конструктор для создания нового экземпляра AuthController
func NewAuthController(authUsecase AuthUseCase, logger Logger) *AuthController {
	if authUsecase == nil {
		log.Fatalf("AuthController initialization failed: authUsecase is nil")
	}
	if logger == nil {
		log.Fatalf("AuthController initialization failed: logger is nil")
	}
	return &AuthController{authUsecase: authUsecase, logger: logger}
}

// DummyLogin обрабатывает запрос на фиктивный логин
func (c *AuthController) DummyLogin(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "AuthController", "method", "DummyLogin", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	var req onlymodels.PostDummyLoginJSONBody
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidRole})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	token, err := c.authUsecase.DummyLogin(contWithTimeout, &req)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.JSON(token)
}

// Register обрабатывает запрос на регистрацию пользователя
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "AuthController", "method", "Register", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	var req onlymodels.PostRegisterJSONBody
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidRequest})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	user, err := c.authUsecase.Register(contWithTimeout, &req)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(user)
}

// Login обрабатывает запрос на логин пользователя
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "AuthController", "method", "Login", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	var req onlymodels.PostLoginJSONBody
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidRequest})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	token, err := c.authUsecase.Login(contWithTimeout, &req)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.JSON(token)
}
