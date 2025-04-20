// Package initdb это функции для инициализации и закрытия БД
package initdb

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"time"

	//драйвер
	_ "github.com/lib/pq"
)

// ConnectionConfig -
type ConnectionConfig interface {
	GetHost() string
	GetDbPort() string
	GetUser() string
	GetPassword() string
	GetDbName() string
}

// NewDBConnection конструктор
func NewDBConnection(cfg ConnectionConfig) (*sqlx.DB, error) {
	if cfg == nil {
		log.Fatalf("nil migration config")
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.GetHost(), cfg.GetDbPort(), cfg.GetUser(), cfg.GetPassword(), cfg.GetDbName(),
	)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
