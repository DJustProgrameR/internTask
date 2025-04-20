// Package repository это интерфейсы репозиториев
package repository

import (
	"context"
	"internshipPVZ/internal/domain/repository/dao"
)

// UserRepo репозиторий
type UserRepo interface {
	// Create добавляет user id в dao
	Create(ctx context.Context, user *dao.User) error
	FindByEmail(ctx context.Context, email string) (*dao.User, error)
}
