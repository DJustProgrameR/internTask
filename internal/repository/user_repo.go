// Package repository это имплементации репозиториев
package repository

import (
	"context"
	"errors"
	sqrl "github.com/Masterminds/squirrel"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	"log"
)

const (
	errorViolatesUniqueEmailConstraint = "pq: duplicate key value violates unique constraint \"users_email_key\""
)

// UserRepo реализация репозитория для пользователей
type UserRepo struct {
	db *sqrl.StmtCache
	qb sqrl.StatementBuilderType
}

// NewUserRepo конструктор для создания нового экземпляра UserRepo
func NewUserRepo(config Config) *UserRepo {
	if config == nil {
		log.Fatalf("user repo config is nil")
		return nil
	}
	if config.GetDbConnection() == nil {
		log.Fatalf("user repo config.GetDbConnection() is nil")
	}
	return &UserRepo{
		db: sqrl.NewStmtCache(config.GetDbConnection()),
		qb: sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar),
	}
}

// Create добавляет нового пользователя в базу данных
func (r *UserRepo) Create(ctx context.Context, user *dao.User) error {
	err := r.qb.Insert("users").
		Columns("email", "password", "role").
		Values(user.Email, user.Password, user.Role).
		Suffix("RETURNING id").
		RunWith(r.db).
		QueryRowContext(ctx).Scan(&user.ID)
	if err != nil {
		if err.Error() == errorViolatesUniqueEmailConstraint {
			return errors.New(model.ErrUserWithEmailAlreadyExists)
		}
	}
	return err
}

// FindByEmail находит пользователя по email
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*dao.User, error) {
	row := r.qb.Select("id", "email", "password", "role").
		From("users").
		Where(sqrl.Eq{"email": email}).
		RunWith(r.db).QueryRowContext(ctx)

	u := &dao.User{}
	err := row.Scan(&u.ID, &u.Email, &u.Password, &u.Role)
	if err != nil {
		if err.Error() == errorNoSQLRows {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}
