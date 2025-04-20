// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/http/onlymodels"
	"log"
)

// NewUseCaseCreatePVZ конструктор
func NewUseCaseCreatePVZ(
	pvzRepo repo.PVZRepo,
	timeService TimeService,
	logger Logger,
) *CreatePVZ {
	if pvzRepo == nil {
		log.Fatalf("CreatePVZ usecase pvzRepo nil")

	}
	if timeService == nil {
		log.Fatalf("CreatePVZ usecase timeService nil")

	}
	if logger == nil {
		log.Fatalf("CreatePVZ usecase logger nil")

	}

	return &CreatePVZ{
		pvzRepo:     pvzRepo,
		timeService: timeService,
		logger:      logger,
	}
}

// CreatePVZ юзкейс
type CreatePVZ struct {
	pvzRepo     repo.PVZRepo
	timeService TimeService
	logger      Logger
}

// Execute создаёт пвз
func (uc *CreatePVZ) Execute(ctx context.Context, request *onlymodels.PVZ, userRole string) (*onlymodels.PVZ, error) {
	pvz, role, err := uc.validateInput(request, userRole)
	if err != nil {
		uc.logger.Error("validation failed",
			"usecase", "CreatePVZ",
			"method", "validateInput",
			"pvz", request,
			"error", err)
		return nil, err
	}

	if model.RoleModerator != role {
		uc.logger.Warn("access denied",
			"usecase", "CreatePVZ",
			"method", "Execute",
			"required_role", model.RoleModerator,
			"user_role", role)
		return nil, errors.New(model.ErrAccessDenied)
	}

	pvzDao := pvz.ToDao()
	pvzDao.RegistrationDate = uc.timeService.GetTime()

	err = uc.pvzRepo.Create(ctx, pvzDao)
	if err != nil {
		uc.logger.Error("failed to create PVZ",
			"usecase", "CreatePVZ",
			"method", "pvzRepo.Create",
			"pvz", pvzDao,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	pvzDto, err := pvzDaoToDto(pvzDao)
	if err != nil {
		uc.logger.Error("failed to convert PVZ DAO to DTO",
			"usecase", "CreatePVZ",
			"method", "pvzDaoToDto",
			"pvz", pvzDao,
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	uc.logger.Info("PVZ created successfully",
		"usecase", "CreatePVZ",
		"pvz_id", pvzDao.ID)
	return pvzDto, nil
}

func (uc *CreatePVZ) validateInput(request *onlymodels.PVZ, userRole string) (pvz *model.PVZ, role model.Role, err error) {
	if role, err = validateRole(userRole); err != nil {
		uc.logger.Warn("invalid user role",
			"usecase", "CreatePVZ",
			"method", "validateInput",
			"user_role", userRole)
		return
	}

	pvz = &model.PVZ{}

	if pvz.City, err = uc.validateCity(string(request.City)); err != nil {
		uc.logger.Warn("invalid city",
			"usecase", "CreatePVZ",
			"method", "validateCity",
			"city", request.City)
		return
	}
	return
}

func (uc *CreatePVZ) validateCity(city string) (model.City, error) {
	switch city {
	case model.CityMoscow.Get():
		return model.CityMoscow, nil
	case model.CitySPB.Get():
		return model.CitySPB, nil
	case model.CityKazan.Get():
		return model.CityKazan, nil
	default:
		return model.CityDefault, errors.New(model.ErrInvalidCityName)
	}
}
