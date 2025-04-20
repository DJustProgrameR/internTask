// Package repository это интерфейсы репозиториев
package repository

import (
	"context"
	"internshipPVZ/internal/domain/repository/dao"
)

// ProductRepo репозиторий
type ProductRepo interface {
	// Add добавляет product id в dao
	Add(ctx context.Context, product *dao.Product) error
	CountProducts(ctx context.Context, receptionID string) (int, error)
	DeleteLastFromReception(ctx context.Context, receptionID string) error
}
