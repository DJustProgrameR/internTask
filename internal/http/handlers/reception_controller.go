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

// OpenReceptionUseCase интерфейс для открытия приёмки
type OpenReceptionUseCase interface {
	Execute(ctx context.Context, PVZID uuid.UUID, userRole string) (*onlymodels.Reception, error)
}

// ReceptionController контроллер для управления приёмками
type ReceptionController struct {
	openReceptionUseCase OpenReceptionUseCase
	logger               Logger
}

// NewReceptionController конструктор для создания нового экземпляра ReceptionController
func NewReceptionController(openReceptionUseCase OpenReceptionUseCase, logger Logger) *ReceptionController {
	if openReceptionUseCase == nil {
		log.Fatalf("ReceptionController initialization failed: openReceptionUseCase is nil")
	}
	if logger == nil {
		log.Fatalf("ReceptionController initialization failed: logger is nil")
	}
	return &ReceptionController{openReceptionUseCase: openReceptionUseCase, logger: logger}
}

// CreateReception обрабатывает запрос на создание приёмки
func (c *ReceptionController) CreateReception(ctx *fiber.Ctx) error {
	if ctx == nil {
		c.logger.Error("received nil ctx", "controller", "ReceptionController", "method", "CreateReception", "error")
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
	}
	userRole := getUserRoleFromContext(ctx)
	var req onlymodels.PostReceptionsJSONBody
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInvalidPVZID})
	}
	contWithTimeout, cancel := context.WithTimeout(context.Background(), cancelContextTime)
	defer cancel()
	reception, err := c.openReceptionUseCase.Execute(contWithTimeout, req.PvzId, userRole)
	if err != nil {
		if err.Error() == model.ErrAccessDenied {
			return ctx.Status(fiber.StatusForbidden).JSON(onlymodels.Error{Message: err.Error()})
		}
		return ctx.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: err.Error()})
	}
	return ctx.Status(fiber.StatusCreated).JSON(reception)
}
