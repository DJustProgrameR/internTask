// Package repository это интерфейсы репозиториев
package repository

import (
	"context"
	"internshipPVZ/internal/domain/repository/dao"
)

// ReceptionRepo репозиторий
type ReceptionRepo interface {
	// Create добавляет reception id в dao
	Create(ctx context.Context, reception *dao.Reception) error
	// FindOpened добавляет reception id в dao если reception существует
	FindOpened(ctx context.Context, reception *dao.Reception) error
	CheckIfExists(ctx context.Context, reception *dao.Reception) (bool, error)
	Close(ctx context.Context, rec *dao.Reception) error
}
