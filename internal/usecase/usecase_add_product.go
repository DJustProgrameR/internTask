// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/domain/repository/dao"
	"internshipPVZ/internal/http/onlymodels"
	"log"
	"time"
)

// Logger логгер
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// TimeService сервис для генерации таймстампов
type TimeService interface {
	GetTime() time.Time
}

// NewUseCaseAddProduct конструктор
func NewUseCaseAddProduct(
	receptionRepo repo.ReceptionRepo,
	productRepo repo.ProductRepo,
	pvzRepo repo.PVZRepo,
	timeService TimeService,
	logger Logger,
) *AddProduct {
	if receptionRepo == nil {
		log.Fatalf("AddProduct usecase receptionRepo nil")
	}
	if productRepo == nil {
		log.Fatalf("AddProduct usecase productRepo nil")
	}
	if pvzRepo == nil {
		log.Fatalf("AddProduct usecase pvzRepo nil")
	}
	if timeService == nil {
		log.Fatalf("AddProduct usecase timeService nil")
	}
	if logger == nil {
		log.Fatalf("AddProduct usecase logger nil")
	}
	return &AddProduct{
		receptionRepo: receptionRepo,
		productRepo:   productRepo,
		pvzRepo:       pvzRepo,
		timeService:   timeService,
		logger:        logger,
	}
}

// AddProduct юзкейс
type AddProduct struct {
	receptionRepo repo.ReceptionRepo
	productRepo   repo.ProductRepo
	pvzRepo       repo.PVZRepo
	timeService   TimeService
	logger        Logger
}

// Execute добавляет продукт
func (uc *AddProduct) Execute(ctx context.Context, request *onlymodels.PostProductsJSONBody, userRole string) (*onlymodels.Product, error) {
	pvzID, role, product, err := uc.validateInput(request, userRole)
	if err != nil {
		uc.logger.Error("validation failed",
			"usecase", "AddProduct",
			"method", "validateInput",
			"error", err)
		return nil, err
	}
	if model.RoleEmployee != role {
		uc.logger.Warn("access denied",
			"usecase", "AddProduct",
			"method", "Execute",
			"required_role", model.RoleEmployee,
			"user_role", role)
		return nil, errors.New(model.ErrAccessDenied)
	}

	product.DateTime = uc.timeService.GetTime()
	rec := &model.Reception{PVZID: pvzID, Status: model.ReceptionInProgress}
	recDao := rec.ToDao()

	exists, err := uc.pvzRepo.CheckIfExists(ctx, recDao.PVZID)
	if err != nil {
		uc.logger.Error("failed to check PVZ existence",
			"usecase", "AddProduct",
			"method", "pvzRepo.CheckIfExists",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}
	if !exists {
		uc.logger.Warn("invalid PVZ ID",
			"usecase", "AddProduct",
			"method", "Execute",
			"pvz_id", recDao.PVZID)
		return nil, errors.New(model.ErrInvalidPVZID)
	}

	recDao.ID = ""
	err = uc.receptionRepo.FindOpened(ctx, recDao)
	if err != nil {
		uc.logger.Error("failed to find opened reception",
			"usecase", "AddProduct",
			"method", "receptionRepo.FindOpened",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}
	if recDao.ID == "" {
		uc.logger.Warn("no active reception found",
			"usecase", "AddProduct",
			"method", "Execute")
		return nil, errors.New(model.ErrNoActiveReception)
	}
	product.ReceptionID, err = validateRawID(recDao.ID)
	if err != nil {
		uc.logger.Error("failed to validate reception ID",
			"usecase", "AddProduct",
			"method", "validateRawID",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}
	productDao := product.ToDao()

	err = uc.productRepo.Add(ctx, productDao)
	if err != nil {
		uc.logger.Error("failed to add product",
			"usecase", "AddProduct",
			"method", "productRepo.Add",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	productDto, err := uc.productDaoToDto(productDao)
	if err != nil {
		uc.logger.Error("failed to convert product DAO to DTO",
			"usecase", "AddProduct",
			"method", "productDaoToDto",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	uc.logger.Info("product added successfully",
		"usecase", "AddProduct",
		"product_id", productDao.ID,
		"reception_id", productDao.ReceptionID)
	return productDto, nil
}

func (uc *AddProduct) validateInput(request *onlymodels.PostProductsJSONBody, userRole string) (pvzID uuid.UUID, role model.Role, product *model.Product, err error) {
	if pvzID, err = validateID(request.PvzId); err != nil {
		return
	}
	if role, err = validateRole(userRole); err != nil {
		return
	}
	product = &model.Product{}
	if product.Type, err = uc.validateProductType(string(request.Type)); err != nil {
		return
	}
	return
}

func (uc *AddProduct) validateProductType(productType string) (model.ProductType, error) {
	switch productType {
	case model.ProductClothes.Get():
		return model.ProductClothes, nil
	case model.ProductElectronics.Get():
		return model.ProductElectronics, nil
	case model.ProductShoes.Get():
		return model.ProductShoes, nil
	default:
		uc.logger.Warn("invalid product type",
			"usecase", "AddProduct",
			"method", "validateProductType",
			"product_type", productType)
		return model.ProductDefault, errors.New(model.ErrInvalidProductType)
	}
}

func (uc *AddProduct) productDaoToDto(product *dao.Product) (*onlymodels.Product, error) {
	if product == nil {
		return nil, errors.New("product is nil")
	}
	id, err := validateRawID(product.ID)
	if err != nil {
		return nil, err
	}
	receptionID, err := validateRawID(product.ReceptionID)
	if err != nil {
		return nil, err
	}
	dto := &onlymodels.Product{
		Id:          &id,
		DateTime:    &product.DateTime,
		Type:        onlymodels.ProductType(model.NewProductType(product.Type)),
		ReceptionId: receptionID,
	}
	return dto, nil
}
