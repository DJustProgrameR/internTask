// Package repository это имплементации репозиториев
package repository

import (
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"internshipPVZ/internal/domain/repository/dao"
	"log"
)

// PvzRepo реализация репозитория для ПВЗ
type PvzRepo struct {
	db *sqrl.StmtCache
	qb sqrl.StatementBuilderType
}

// NewPVZRepo конструктор для создания нового экземпляра PvzRepo
func NewPVZRepo(config Config) *PvzRepo {
	if config == nil {
		log.Fatalf("pvz repo config is nil")
		return nil
	}
	if config.GetDbConnection() == nil {
		log.Fatalf("pvz repo config.GetDbConnection() is nil")
	}
	return &PvzRepo{
		sqrl.NewStmtCache(config.GetDbConnection()),
		sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar),
	}
}

// Create добавляет новый ПВЗ в базу данных
func (r *PvzRepo) Create(ctx context.Context, pvz *dao.PVZ) error {
	err := r.qb.Insert("pvz").
		Columns("registration_date", "city").
		Values(pvz.RegistrationDate, pvz.City).
		Suffix("RETURNING id").
		RunWith(r.db).
		QueryRowContext(ctx).Scan(&pvz.ID)
	return err
}

// GetAllWithFilter возвращает список ПВЗ с фильтрацией по дате и пагинацией
func (r *PvzRepo) GetAllWithFilter(ctx context.Context, startDate, endDate string, page, limit int) ([]*dao.PVZList, error) {
	offset := (page - 1) * limit
	subQuery := r.qb.Select("id").
		From("pvz").
		OrderBy("id").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	pvzQuery := r.qb.Select(
		"p.id",
		"p.registration_date",
		"p.city",
		"r.id",
		"r.date_time",
		"r.status",
		"pr.id",
		"pr.date_time",
		"pr.type",
	).
		From("pvz p").
		InnerJoin("receptions r ON p.id = r.pvz_id").
		LeftJoin("products pr ON r.id = pr.reception_id")
	if len(startDate) > 0 {
		pvzQuery = pvzQuery.Where(sqrl.GtOrEq{"r.date_time": startDate})
	}
	if len(endDate) > 0 {
		pvzQuery = pvzQuery.Where(sqrl.LtOrEq{"r.date_time": endDate})
	}
	pvzQuery = pvzQuery.Where(sqrl.Expr("p.id in (?)", subQuery))

	pvzRows, err := pvzQuery.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	var pvzs []*dao.PVZList
	for pvzRows.Next() {
		pvz := dao.PVZList{}
		err := pvzRows.Scan(
			&pvz.PvzID,
			&pvz.RegistrationDate,
			&pvz.City,
			&pvz.ReceptionID,
			&pvz.ReceptionDateTime,
			&pvz.Status,
			&pvz.ProductID,
			&pvz.ProductDateTime,
			&pvz.Type,
		)
		if err != nil {
			return nil, err
		}
		pvzs = append(pvzs, &pvz)
	}
	if err := pvzRows.Close(); err != nil {
		return nil, err
	}
	return pvzs, nil
}

// Get возвращает список всех ПВЗ
func (r *PvzRepo) Get(ctx context.Context) ([]*dao.PVZ, error) {
	rows, err := r.qb.Select("id", "registration_date", "city").From("pvz").
		RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	var pvzs []*dao.PVZ
	for rows.Next() {
		pvz := dao.PVZ{}
		err := rows.Scan(
			&pvz.ID,
			&pvz.RegistrationDate,
			&pvz.City,
		)
		if err != nil {
			return nil, err
		}
		pvzs = append(pvzs, &pvz)
	}
	return pvzs, nil
}

// CheckIfExists проверяет существование ПВЗ по ID
func (r *PvzRepo) CheckIfExists(ctx context.Context, id string) (bool, error) {
	count := 0
	err := r.qb.Select("count(*)").
		From("pvz").
		Where(sqrl.Eq{"id": id}).
		RunWith(r.db).
		QueryRowContext(ctx).
		Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 1, nil
}
