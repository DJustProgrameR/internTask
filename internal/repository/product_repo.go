// Package repository это имплементации репозиториев
package repository

import (
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
	"internshipPVZ/internal/domain/repository/dao"
)

// ProductRepo реализация репозитория для продуктов
type ProductRepo struct {
	db *sqrl.StmtCache
	qb sqrl.StatementBuilderType
}

// Config интерфейс конфигурации репозитория
type Config interface {
	GetDbConnection() *sqlx.DB
}

// NewProductRepo конструктор для создания нового экземпляра ProductRepo
func NewProductRepo(config Config) *ProductRepo {
	if config == nil {
		log.Fatalf("product repo config is nil")
		return nil
	}
	if config.GetDbConnection() == nil {
		log.Fatalf("product repo config.GetDbConnection() is nil")
	}
	return &ProductRepo{
		sqrl.NewStmtCache(config.GetDbConnection()),
		sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar),
	}
}

// Add добавляет новый продукт в базу данных
func (r *ProductRepo) Add(ctx context.Context, product *dao.Product) error {
	err := r.qb.Insert("products").
		Columns("reception_id", "date_time", "type").
		Values(product.ReceptionID, product.DateTime, product.Type).
		Suffix("RETURNING id").
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&product.ID)
	return err
}

// CountProducts возвращает количество продуктов для данной приёмки
func (r *ProductRepo) CountProducts(ctx context.Context, receptionID string) (int, error) {
	count := 0
	err := r.qb.Select("count(*)").
		From("products").
		Where(sqrl.Eq{"reception_id": receptionID}).
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteLastFromReception удаляет последний добавленный продукт из приёмки
func (r *ProductRepo) DeleteLastFromReception(ctx context.Context, receptionID string) error {
	subQuery := r.qb.Select("id").
		From("products").
		Where(sqrl.Eq{"reception_id": receptionID}).
		OrderBy("date_time DESC").
		Limit(1)
	_, err := r.qb.Delete("products").
		Where(sqrl.Expr("id IN (?)", subQuery)).
		RunWith(r.db).ExecContext(ctx)
	return err
}
