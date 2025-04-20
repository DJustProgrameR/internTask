// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/http/onlymodels"
	"log"
)

// NewUseCaseCloseReception конструктор
func NewUseCaseCloseReception(
	receptionRepo repo.ReceptionRepo,
	pvzRepo repo.PVZRepo,
	logger Logger,
) *CloseReception {
	if receptionRepo == nil {
		log.Fatalf("CloseReception usecase receptionRepo nil")

	}
	if pvzRepo == nil {
		log.Fatalf("CloseReception usecase pvzRepo nil")

	}
	if logger == nil {
		log.Fatalf("CloseReception usecase logger nil")

	}

	return &CloseReception{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		logger:        logger,
	}
}

// CloseReception юзкейс
type CloseReception struct {
	receptionRepo repo.ReceptionRepo
	pvzRepo       repo.PVZRepo
	logger        Logger
}

// Execute закрывает приёмку
func (uc *CloseReception) Execute(ctx context.Context, PVZID uuid.UUID, userRole string) (*onlymodels.Reception, error) {
	pvzID, role, err := uc.validateInput(PVZID, userRole)
	if err != nil {
		uc.logger.Error("validation failed",
			"usecase", "CloseReception",
			"method", "validateInput",
			"pvz_id", PVZID,
			"error", err)
		return nil, err
	}

	if model.RoleEmployee != role {
		uc.logger.Warn("access denied",
			"usecase", "CloseReception",
			"method", "Execute",
			"required_role", model.RoleEmployee,
			"user_role", role)
		return nil, errors.New(model.ErrAccessDenied)
	}

	rec := &model.Reception{PVZID: pvzID, Status: model.ReceptionInProgress}
	recDao := rec.ToDao()

	exists, err := uc.pvzRepo.CheckIfExists(ctx, recDao.PVZID)
	if err != nil {
		uc.logger.Error("failed to check PVZ existence",
			"usecase", "CloseReception",
			"method", "pvzRepo.CheckIfExists",
			"pvz_id", recDao.PVZID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	if !exists {
		uc.logger.Warn("invalid PVZ ID",
			"usecase", "CloseReception",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return nil, errors.New(model.ErrInvalidPVZID)
	}

	recDao.ID = ""
	err = uc.receptionRepo.FindOpened(ctx, recDao)
	if err != nil {
		uc.logger.Error("failed to find opened reception",
			"usecase", "CloseReception",
			"method", "receptionRepo.FindOpened",
			"pvz_id", recDao.PVZID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	if recDao.ID == "" {
		uc.logger.Warn("no active reception found",
			"usecase", "CloseReception",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return nil, errors.New(model.ErrNoActiveReception)
	}

	recDao.Status = model.ReceptionClosed.ToInt()

	err = uc.receptionRepo.Close(ctx, recDao)
	if err != nil {
		uc.logger.Error("failed to close reception",
			"usecase", "CloseReception",
			"method", "receptionRepo.Close",
			"reception_id", recDao.ID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	recDto, err := receptionDaoToDto(recDao)
	if err != nil {
		uc.logger.Error("failed to convert reception DAO to DTO",
			"usecase", "CloseReception",
			"method", "receptionDaoToDto",
			"reception_id", recDao.ID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	uc.logger.Info("reception closed successfully",
		"usecase", "CloseReception",
		"reception_id", recDao.ID,
		"pvz_id", recDao.PVZID)
	return recDto, nil
}

func (uc *CloseReception) validateInput(PVZID uuid.UUID, userRole string) (pvzID uuid.UUID, role model.Role, err error) {
	if pvzID, err = validateID(PVZID); err != nil {
		uc.logger.Warn("invalid PVZ ID format",
			"usecase", "CloseReception",
			"method", "validateInput",
			"pvz_id", PVZID)
		return
	}
	if role, err = validateRole(userRole); err != nil {
		uc.logger.Warn("invalid user role",
			"usecase", "CloseReception",
			"method", "validateInput",
			"user_role", userRole)
		return
	}
	return
}
