// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	"log"
)

// NewUseCaseDeleteProduct конструктор
func NewUseCaseDeleteProduct(
	productRepo repo.ProductRepo,
	receptionRepo repo.ReceptionRepo,
	pvzRepo repo.PVZRepo,
	logger Logger,
) *DeleteProduct {
	if productRepo == nil {
		log.Fatalf("DeleteProduct usecase productRepo nil")

	}
	if receptionRepo == nil {
		log.Fatalf("DeleteProduct usecase receptionRepo nil")

	}
	if pvzRepo == nil {
		log.Fatalf("DeleteProduct usecase pvzRepo nil")

	}
	if logger == nil {
		log.Fatalf("DeleteProduct usecase logger nil")

	}

	return &DeleteProduct{
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
		pvzRepo:       pvzRepo,
		logger:        logger,
	}
}

// DeleteProduct юзкейс
type DeleteProduct struct {
	productRepo   repo.ProductRepo
	receptionRepo repo.ReceptionRepo
	pvzRepo       repo.PVZRepo
	logger        Logger
}

// Execute удаляет продукт
func (uc *DeleteProduct) Execute(ctx context.Context, PVZID uuid.UUID, userRole string) error {
	pvzID, role, err := uc.validateInput(PVZID, userRole)
	if err != nil {
		uc.logger.Error("validation failed",
			"usecase", "DeleteProduct",
			"method", "validateInput",
			"pvz_id", PVZID,
			"error", err)
		return err
	}

	if model.RoleEmployee != role {
		uc.logger.Warn("access denied",
			"usecase", "DeleteProduct",
			"method", "Execute",
			"required_role", model.RoleEmployee,
			"user_role", role)
		return errors.New(model.ErrAccessDenied)
	}

	rec := &model.Reception{PVZID: pvzID, Status: model.ReceptionInProgress}
	recDao := rec.ToDao()

	exists, err := uc.pvzRepo.CheckIfExists(ctx, recDao.PVZID)
	if err != nil {
		uc.logger.Error("failed to check PVZ existence",
			"usecase", "DeleteProduct",
			"method", "pvzRepo.CheckIfExists",
			"pvz_id", recDao.PVZID,
			"error", err)
		return errors.New(model.ErrInternal)
	}

	if !exists {
		uc.logger.Warn("invalid PVZ ID",
			"usecase", "DeleteProduct",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return errors.New(model.ErrInvalidPVZID)
	}

	recDao.ID = ""
	err = uc.receptionRepo.FindOpened(ctx, recDao)
	if err != nil {
		uc.logger.Error("failed to find opened reception",
			"usecase", "DeleteProduct",
			"method", "receptionRepo.FindOpened",
			"pvz_id", recDao.PVZID,
			"error", err)
		return errors.New(model.ErrInternal)
	}

	if recDao.ID == "" {
		uc.logger.Warn("no active reception found",
			"usecase", "DeleteProduct",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return errors.New(model.ErrNoActiveReception)
	}

	_, err = validateRawID(recDao.ID)
	if err != nil {
		uc.logger.Error("invalid reception ID format",
			"usecase", "DeleteProduct",
			"method", "validateRawID",
			"reception_id", recDao.ID,
			"error", err)
		return errors.New(model.ErrInternal)
	}

	numOfProducts, err := uc.productRepo.CountProducts(ctx, recDao.ID)
	if err != nil {
		uc.logger.Error("failed to count products",
			"usecase", "DeleteProduct",
			"method", "productRepo.CountProducts",
			"reception_id", recDao.ID,
			"error", err)
		return errors.New(model.ErrInternal)
	}

	if numOfProducts == 0 {
		uc.logger.Warn("no products left to delete",
			"usecase", "DeleteProduct",
			"method", "Execute",
			"reception_id", recDao.ID)
		return errors.New(model.ErrNoProductsLeftToDelete)
	}

	err = uc.productRepo.DeleteLastFromReception(ctx, recDao.ID)
	if err != nil {
		uc.logger.Error("failed to delete last product from reception",
			"usecase", "DeleteProduct",
			"method", "productRepo.DeleteLastFromReception",
			"reception_id", recDao.ID,
			"error", err)
		return errors.New(model.ErrInternal)
	}

	uc.logger.Info("product deleted successfully",
		"usecase", "DeleteProduct",
		"reception_id", recDao.ID,
		"pvz_id", recDao.PVZID)
	return nil
}

func (uc *DeleteProduct) validateInput(PVZID uuid.UUID, userRole string) (pvzID uuid.UUID, role model.Role, err error) {
	if pvzID, err = validateID(PVZID); err != nil {
		uc.logger.Warn("invalid PVZ ID format",
			"usecase", "DeleteProduct",
			"method", "validateInput",
			"pvz_id", PVZID)
		return
	}

	if role, err = validateRole(userRole); err != nil {
		uc.logger.Warn("invalid user role",
			"usecase", "DeleteProduct",
			"method", "validateInput",
			"user_role", userRole)
		return
	}
	return
}
