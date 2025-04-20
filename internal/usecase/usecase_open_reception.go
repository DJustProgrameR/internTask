// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/http/onlymodels"
	"log"
)

// NewUseCaseOpenReception конструктор
func NewUseCaseOpenReception(
	receptionRepo repo.ReceptionRepo,
	pvzRepo repo.PVZRepo,
	timeService TimeService,
	logger Logger,
) *OpenReception {
	if receptionRepo == nil {
		log.Fatalf("OpenReception usecase receptionRepo nil")

	}
	if pvzRepo == nil {
		log.Fatalf("OpenReception usecase pvzRepo nil")

	}
	if timeService == nil {
		log.Fatalf("OpenReception usecase timeService nil")

	}
	if logger == nil {
		log.Fatalf("OpenReception usecase logger nil")

	}

	return &OpenReception{
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		timeService:   timeService,
		logger:        logger,
	}
}

// OpenReception юзкейс
type OpenReception struct {
	receptionRepo repo.ReceptionRepo
	timeService   TimeService
	pvzRepo       repo.PVZRepo
	logger        Logger
}

// Execute открывает приёмку
func (uc *OpenReception) Execute(ctx context.Context, PVZID uuid.UUID, userRole string) (*onlymodels.Reception, error) {
	pvzID, role, err := uc.validateInput(PVZID, userRole)
	if err != nil {
		uc.logger.Error("validation failed",
			"usecase", "OpenReception",
			"method", "validateInput",
			"pvz_id", PVZID,
			"error", err)
		return nil, err
	}

	if model.RoleEmployee != role {
		uc.logger.Warn("access denied",
			"usecase", "OpenReception",
			"method", "Execute",
			"required_role", model.RoleEmployee,
			"user_role", role)
		return nil, errors.New(model.ErrAccessDenied)
	}

	rec := &model.Reception{PVZID: pvzID, DateTime: uc.timeService.GetTime(), Status: model.ReceptionInProgress}
	recDao := rec.ToDao()

	exists, err := uc.pvzRepo.CheckIfExists(ctx, recDao.PVZID)
	if err != nil {
		uc.logger.Error("failed to check PVZ existence",
			"usecase", "OpenReception",
			"method", "pvzRepo.CheckIfExists",
			"pvz_id", recDao.PVZID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	if !exists {
		uc.logger.Warn("invalid PVZ ID",
			"usecase", "OpenReception",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return nil, errors.New(model.ErrInvalidPVZID)
	}

	exists, err = uc.receptionRepo.CheckIfExists(ctx, recDao)
	if err != nil {
		uc.logger.Error("failed to check if reception exists",
			"usecase", "OpenReception",
			"method", "receptionRepo.CheckIfExists",
			"pvz_id", recDao.PVZID,
			"error", err)
		return nil, fmt.Errorf(model.ErrInternal)
	}

	if exists {
		uc.logger.Warn("reception already opened",
			"usecase", "OpenReception",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return nil, errors.New(model.ErrReceptionAlreadyOpened)
	}

	err = uc.receptionRepo.Create(ctx, recDao)
	if err != nil {
		uc.logger.Error("failed to create reception",
			"usecase", "OpenReception",
			"method", "receptionRepo.Create",
			"pvz_id", recDao.PVZID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	recDto, err := receptionDaoToDto(recDao)
	if err != nil {
		uc.logger.Error("failed to convert reception DAO to DTO",
			"usecase", "OpenReception",
			"method", "receptionDaoToDto",
			"reception_id", recDao.ID,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	uc.logger.Info("reception opened successfully",
		"usecase", "OpenReception",
		"reception_id", recDao.ID,
		"pvz_id", recDao.PVZID)
	return recDto, nil
}

func (uc *OpenReception) validateInput(PVZID uuid.UUID, userRole string) (pvzID uuid.UUID, role model.Role, err error) {
	if pvzID, err = validateID(PVZID); err != nil {
		uc.logger.Warn("invalid PVZ ID format",
			"usecase", "OpenReception",
			"method", "validateInput",
			"pvz_id", PVZID)
		return
	}

	if role, err = validateRole(userRole); err != nil {
		uc.logger.Warn("invalid user role",
			"usecase", "OpenReception",
			"method", "validateInput",
			"user_role", userRole)
		return
	}
	return
}
