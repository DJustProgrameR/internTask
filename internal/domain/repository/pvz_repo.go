// Package repository это интерфейсы репозиториев
package repository

import (
	"context"
	"internshipPVZ/internal/domain/repository/dao"
)

// PVZRepo репозиторий
type PVZRepo interface {
	// Create добавляет pvz id в dao
	Create(ctx context.Context, pvz *dao.PVZ) error
	GetAllWithFilter(ctx context.Context, startDate, endDate string, page, limit int) ([]*dao.PVZList, error)
	Get(ctx context.Context) ([]*dao.PVZ, error)
	CheckIfExists(ctx context.Context, id string) (bool, error)
}
