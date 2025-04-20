// Package handlers это http хэндлеры
package handlers

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/http/onlymodels"
	"log"
)

// CreatePVZUseCase интерфейс для создания ПВЗ
type CreatePVZUseCase interface {
	Execute(ctx context.Context, request *onlymodels.PVZ, userRole string) (*onlymodels.PVZ, error)
}

// DeleteProductUseCase интерфейс для удаления продукта
type DeleteProductUseCase interface {
	Execute(ctx context.Context, PVZID uuid.UUID, userRole string) error
}

// CloseReceptionUseCase интерфейс для закрытия приёмки
type CloseReceptionUseCase interface {
	Execute(ctx context.Context, PVZID uuid.UUID, userRole string) (*onlymodels.Reception, error)
}

// GetPVZUseCase интерфейс для получения списка ПВЗ
type GetPVZUseCase interface {
	GetFiltered(ctx context.Context, startDate, endDate string, page, limit int, userRole string) *onlymodels.GetFilteredResponse
}

// PVZController контроллер для управления ПВЗ
type PVZController struct {
	closeReceptionUseCase CloseReceptionUseCase
	deleteProductUseCase  DeleteProductUseCase
	createPVZUseCase      CreatePVZUseCase
	getPVZUseCase         GetPVZUseCase
	logger                Logger
}

// NewPVZController конструктор для создания нового экземпляра PVZController
func NewPVZController(
	deleteProductUseCase DeleteProductUseCase,
	createPVZUseCase CreatePVZUseCase,
	closeReceptionUseCase CloseReceptionUseCase,
	getPVZUseCase GetPVZUseCase,
	logger Logger,
) *PVZController {
	if deleteProductUseCase == nil {
		log.Fatalf("PVZController initialization failed: deleteProductUseCase is nil")
	}
	if createPVZUseCase == nil {
		log.Fatalf("PVZController initialization failed: createPVZUseCase is nil")
	}
	if closeReceptionUseCase == nil {
		log.Fatalf("PVZController initialization failed: closeReceptionUseCase is nil")
	}
	if getPVZUseCase == nil {
		log.Fatalf("PVZController initialization failed: getPVZUseCase is nil")
	}
	if logger == nil {
		log.Fatalf("PVZController initialization failed: logger is nil")
	}
	return &PVZController{
		deleteProductUseCase:  deleteProductUseCase,
		createPVZUseCase:      createPVZUseCase,
		closeReceptionUseCase: closeReceptionUseCase,
		getPVZUseCase:         getPVZUseCase,
		logger:                logger,
	}
}

// CreatePVZ обрабатывает запрос на создание ПВЗ
func (c *PVZController) CreatePVZ(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "PVZController", "method", "CreatePVZ", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	userRole := getUserRoleFromContext(ctx)
	var req onlymodels.PVZ
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidRequest})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	pvz, err := c.createPVZUseCase.Execute(contWithTimeout, &req, userRole)
	if err != nil {
		if err.Error() == model.ErrAccessDenied {
			return ctx.Status(fiber.StatusForbidden).JSON(onlymodels.Error{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(pvz)
}

// GetPVZs обрабатывает запрос на получение списка ПВЗ
func (c *PVZController) GetPVZs(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "PVZController", "method", "GetPVZs", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	userRole := getUserRoleFromContext(ctx)
	startDate := ctx.Query("startDate")
	endDate := ctx.Query("endDate")
	page := ctx.QueryInt("page")
	limit := ctx.QueryInt("limit")
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	pvzs := c.getPVZUseCase.GetFiltered(contWithTimeout, startDate, endDate, page, limit, userRole)
	return ctx.JSON(pvzs)
}

// CloseLastReception обрабатывает запрос на закрытие последней приёмки
func (c *PVZController) CloseLastReception(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "PVZController", "method", "CloseLastReception", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	userRole := getUserRoleFromContext(ctx)
	pvzID, err := uuid.Parse(ctx.Params("pvzId"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidPVZID})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	reception, err := c.closeReceptionUseCase.Execute(contWithTimeout, pvzID, userRole)
	if err != nil {
		if err.Error() == model.ErrAccessDenied {
			return ctx.Status(fiber.StatusForbidden).JSON(onlymodels.Error{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.JSON(reception)
}

// DeleteLastProduct обрабатывает запрос на удаление последнего продукта
func (c *PVZController) DeleteLastProduct(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "PVZController", "method", "DeleteLastProduct", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	userRole := getUserRoleFromContext(ctx)
	pvzID, err := uuid.Parse(ctx.Params("pvzId"))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidPVZID})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	err = c.deleteProductUseCase.Execute(contWithTimeout, pvzID, userRole)
	if err != nil {
		if err.Error() == model.ErrAccessDenied {
			return ctx.Status(fiber.StatusForbidden).JSON(onlymodels.Error{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.SendStatus(fiber.StatusOK)
}
