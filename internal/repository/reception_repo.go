// Package repository это имплементации репозиториев
package repository

import (
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"internshipPVZ/internal/domain/repository/dao"
	"log"
)

const (
	errorNoSQLRows string = "sql: no rows in result set"
)

// ReceptionRepo реализация репозитория для приёмок
type ReceptionRepo struct {
	db *sqrl.StmtCache
	qb sqrl.StatementBuilderType
}

// NewReceptionRepo конструктор для создания нового экземпляра ReceptionRepo
func NewReceptionRepo(config Config) *ReceptionRepo {
	if config == nil {
		log.Fatalf("reception repo config is nil")
		return nil
	}
	if config.GetDbConnection() == nil {
		log.Fatalf("reception repo config.GetDbConnection() is nil")
	}
	return &ReceptionRepo{
		sqrl.NewStmtCache(config.GetDbConnection()),
		sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar),
	}
}

// Create добавляет новую приёмку в базу данных
func (r *ReceptionRepo) Create(ctx context.Context, rec *dao.Reception) error {
	err := r.qb.Insert("receptions").
		Columns("pvz_id", "date_time", "status").
		Values(rec.PVZID, rec.DateTime, rec.Status).
		Suffix("RETURNING id").
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&rec.ID)
	return err
}

// FindOpened находит открытую приёмку для данного ПВЗ
func (r *ReceptionRepo) FindOpened(ctx context.Context, rec *dao.Reception) error {
	err := r.qb.Select("id").
		From("receptions").
		Where(sqrl.And{
			sqrl.Eq{"pvz_id": rec.PVZID},
			sqrl.Eq{"status": rec.Status},
		}).
		Limit(1).
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&rec.ID)
	if err != nil {
		if err.Error() == errorNoSQLRows {
			return nil
		}
		return err
	}
	return err
}

// CheckIfExists проверяет существование приёмки для данного ПВЗ
func (r *ReceptionRepo) CheckIfExists(ctx context.Context, rec *dao.Reception) (bool, error) {
	count := 0
	err := r.qb.Select("count(*)").
		From("receptions").
		Where(sqrl.And{
			sqrl.Eq{"pvz_id": rec.PVZID},
			sqrl.Eq{"status": rec.Status},
		}).
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}

// Close закрывает приёмку по её ID
func (r *ReceptionRepo) Close(ctx context.Context, rec *dao.Reception) error {
	_, err := r.qb.Update("receptions").
		Set("status", rec.Status).
		Where(sqrl.Eq{"id": rec.ID}).
		RunWith(r.db).ExecContext(ctx)
	return err
}
