// Package config это конфиг приложения
package config

import (
	// для парсинга пути к миграциям
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"os"
)

// NewAppConfig конструктор
func NewAppConfig() *AppConfig {
	return &AppConfig{
		appPort:   os.Getenv("SERVER_PORT"),
		logLevel:  os.Getenv("LOG_LEVEL"),
		jwtSecret: os.Getenv("JWT_SECRET"),
		db: &DB{
			host:         os.Getenv("DATABASE_HOST"),
			port:         os.Getenv("DATABASE_PORT"),
			user:         os.Getenv("DATABASE_USER"),
			password:     os.Getenv("DATABASE_PASSWORD"),
			name:         os.Getenv("DATABASE_NAME"),
			migrationDir: "file://" + os.Getenv("MIGRATIONS_DIR"),
		},
		prometheus: &Prometheus{
			port: os.Getenv("PROM_PORT"),
		},
		grpc: &GRPC{
			port: os.Getenv("GRPC_PORT"),
		},
	}
}

// Prometheus конфиг
type Prometheus struct {
	port string
}

// GRPC конфиг
type GRPC struct {
	port string
}

// DB конфиг
type DB struct {
	host         string
	port         string
	user         string
	password     string
	name         string
	connection   *sqlx.DB
	migrationDir string
}

// AppConfig конфиг
type AppConfig struct {
	appPort    string
	logLevel   string
	db         *DB
	grpc       *GRPC
	jwtSecret  string
	prometheus *Prometheus
}

// GetAppPort возвращает порт приложения.
func (ac *AppConfig) GetAppPort() string {
	return ac.appPort
}

// GetLogLevel возвращает уровень логирования.
func (ac *AppConfig) GetLogLevel() string {
	return ac.logLevel
}

// GetHost возвращает хост базы данных.
func (ac *AppConfig) GetHost() string {
	return ac.db.host
}

// GetDbPort возвращает порт базы данных.
func (ac *AppConfig) GetDbPort() string {
	return ac.db.port
}

// GetGrpcPort возвращает порт gRPC.
func (ac *AppConfig) GetGrpcPort() string {
	return ac.grpc.port
}

// GetPrometheusPort возвращает порт Prometheus.
func (ac *AppConfig) GetPrometheusPort() string {
	return ac.prometheus.port
}

// GetUser возвращает имя пользователя базы данных.
func (ac *AppConfig) GetUser() string {
	return ac.db.user
}

// GetPassword возвращает пароль базы данных.
func (ac *AppConfig) GetPassword() string {
	return ac.db.password
}

// GetDbName возвращает имя базы данных.
func (ac *AppConfig) GetDbName() string {
	return ac.db.name
}

// SetDbConnection устанавливает соединение с базой данных.
func (ac *AppConfig) SetDbConnection(conn *sqlx.DB) {
	ac.db.connection = conn
}

// GetDbConnection возвращает соединение с базой данных.
func (ac *AppConfig) GetDbConnection() *sqlx.DB {
	return ac.db.connection
}

// GetMigrationDir возвращает директорию миграций.
func (ac *AppConfig) GetMigrationDir() string {
	return ac.db.migrationDir
}

// GetJWTSecret возвращает секретный ключ JWT.
func (ac *AppConfig) GetJWTSecret() string {
	return ac.jwtSecret
}
