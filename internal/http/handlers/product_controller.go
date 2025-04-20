// Package handlers это http хэндлеры
package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/http/onlymodels"
	"log"
)

// AddProductUsecase интерфейс для добавления продукта
type AddProductUsecase interface {
	Execute(ctx context.Context, request *onlymodels.PostProductsJSONBody, userRole string) (*onlymodels.Product, error)
}

// ProductController контроллер для управления продуктами
type ProductController struct {
	addProductUsecase AddProductUsecase
	logger            Logger
}

// NewProductController конструктор для создания нового экземпляра ProductController
func NewProductController(addProductUsecase AddProductUsecase, logger Logger) *ProductController {
	if addProductUsecase == nil {
		log.Fatalf("ProductController initialization failed: addProductUsecase is nil")
	}
	if logger == nil {
		log.Fatalf("ProductController initialization failed: logger is nil")
	}
	return &ProductController{addProductUsecase: addProductUsecase, logger: logger}
}

// CreateProduct обрабатывает запрос на создание продукта
func (c *ProductController) CreateProduct(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "ProductController", "method", "CreateProduct", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	UserRole := getUserRoleFromContext(ctx)
	var req onlymodels.PostProductsJSONBody
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidPVZID})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()

	product, err := c.addProductUsecase.Execute(contWithTimeout, &req, UserRole)
	if err != nil {
		if err.Error() == model.ErrAccessDenied {
			return ctx.Status(fiber.StatusForbidden).JSON(onlymodels.Error{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}

	return ctx.Status(fiber.StatusCreated).JSON(product)
}
