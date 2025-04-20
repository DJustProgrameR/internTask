// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	pb "internshipPVZ/internal/grpc/models"
	"internshipPVZ/internal/http/onlymodels"
	"log"
	"strings"
	"time"
)

// NewUseCaseGetPvz конструктор
func NewUseCaseGetPvz(
	pvzRepo repo.PVZRepo,
	logger Logger,
) *GetPvz {
	if pvzRepo == nil {
		log.Fatalf("GetPvz usecase pvzRepo nil")

	}
	if logger == nil {
		log.Fatalf("GetPvz usecase logger nil")

	}

	return &GetPvz{
		pvzRepo: pvzRepo,
		logger:  logger,
	}
}

// GetPvz юзкейс
type GetPvz struct {
	pvzRepo repo.PVZRepo
	logger  Logger
}

// GetFiltered выдаёт фильтрованные пвз со всей информацией
func (uc *GetPvz) GetFiltered(ctx context.Context, startDate, endDate string, pageR, limitR int, userRole string) *onlymodels.GetFilteredResponse {
	page, limit, role, err := uc.validateInput(pageR, limitR, userRole)
	if err != nil {
		uc.logger.Error("validation failed",
			"usecase", "GetPvz",
			"method", "validateInput",
			"page", pageR,
			"limit", limitR,
			"user_role", userRole,
			"error", err)
		return nil
	}

	if model.RoleEmployee != role && model.RoleModerator != role {
		uc.logger.Warn("access denied",
			"usecase", "GetPvz",
			"method", "GetFiltered",
			"required_role", "RoleEmployee or RoleModerator",
			"user_role", role)
		return nil
	}

	if startDate != "" {
		_, err := uc.validateTime(startDate)
		if err != nil {
			uc.logger.Warn("invalid start date format",
				"usecase", "GetPvz",
				"method", "validateTime",
				"start_date", startDate)
			return nil
		}
		startDate = strings.Join([]string{startDate, " 00:00:00"}, "")
	}

	if endDate != "" {
		_, err := uc.validateTime(endDate)
		if err != nil {
			uc.logger.Warn("invalid end date format",
				"usecase", "GetPvz",
				"method", "validateTime",
				"end_date", endDate)
			return nil
		}
		endDate = strings.Join([]string{endDate, " 23:59:59"}, "")
	}

	response, err := uc.pvzRepo.GetAllWithFilter(ctx, startDate, endDate, page, limit)
	if err != nil {
		uc.logger.Error("failed to get filtered PVZs",
			"usecase", "GetPvz",
			"method", "pvzRepo.GetAllWithFilter",
			"start_date", startDate,
			"end_date", endDate,
			"page", page,
			"limit", limit,
			"error", err)
		return nil
	}

	pvzFiltered, err := pvzListToDto(response)
	if err != nil {
		uc.logger.Error("failed to convert PVZ list to DTO",
			"usecase", "GetPvz",
			"method", "pvzListToDto",
			"error", err)
		return nil
	}

	uc.logger.Info("filtered PVZs retrieved successfully",
		"usecase", "GetPvz",
		"start_date", startDate,
		"end_date", endDate,
		"page", page,
		"limit", limit)
	return pvzFiltered
}

// Get выдаёт все пвз с базовой информацией
func (uc *GetPvz) Get(ctx context.Context) *pb.GetPVZListResponse {
	repoResponse, err := uc.pvzRepo.Get(ctx)
	if err != nil {
		uc.logger.Error("failed to get PVZs",
			"usecase", "GetPvz",
			"method", "pvzRepo.Get",
			"error", err)
		return nil
	}

	pvzs, err := pvzsToGrpcDto(repoResponse)
	if err != nil {
		uc.logger.Error("failed to get PVZs",
			"usecase", "GetPvz",
			"method", "pvzsToGrpcDto",
			"error", err)
		return nil
	}
	response := &pb.GetPVZListResponse{Pvzs: pvzs}

	uc.logger.Info("PVZs retrieved successfully",
		"usecase", "GetPvz")
	return response
}

func (uc *GetPvz) validateInput(pageR, limitR int, userRole string) (page, limit int, role model.Role, err error) {
	if role, err = validateRole(userRole); err != nil {
		uc.logger.Warn("invalid user role",
			"usecase", "GetPvz",
			"method", "validateInput",
			"user_role", userRole)
		return
	}

	if pageR == 0 {
		pageR = 1
	}
	page = pageR

	if limitR < 1 || limitR > 30 {
		limitR = 10
	}
	limit = limitR

	return
}

func (uc *GetPvz) validateTime(rTime string) (time.Time, error) {
	rTime = strings.Join([]string{rTime, "T00:00:00Z"}, "")
	dateTime, err := time.Parse(time.RFC3339, rTime)
	if err != nil {
		uc.logger.Warn("invalid time format",
			"usecase", "GetPvz",
			"method", "validateTime",
			"time", rTime)
		return time.Now(), err
	}
	return dateTime, err
}
