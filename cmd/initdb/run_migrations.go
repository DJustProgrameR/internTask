// Package initdb это функции для инициализации и закрытия БД
package initdb

import (
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jmoiron/sqlx"
	"log"
)

// MigrationConfig -
type MigrationConfig interface {
	GetMigrationDir() string
	GetDbName() string
	GetDbConnection() *sqlx.DB
}

// RunMigrations запускает миграции
func RunMigrations(cfg MigrationConfig) error {
	if cfg == nil {
		log.Fatalf("nil migration config")
	}
	if cfg.GetDbConnection() == nil {
		log.Fatalf("nil migration connection")
	}
	if cfg.GetDbConnection().DB == nil {
		log.Fatalf("nil migration connection.DB")
	}
	driver, err := postgres.WithInstance(cfg.GetDbConnection().DB, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		cfg.GetMigrationDir(),
		cfg.GetDbName(),
		driver,
	)

	if err != nil {
		return err
	}
	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
