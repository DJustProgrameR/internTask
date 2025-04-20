// Package integration это интеграционное тестирование
package integration

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

// NewTestAppConfig конструктор
func NewTestAppConfig() *TestAppConfig {
	return &TestAppConfig{
		logLevel:  "dev",
		jwtSecret: "07jvv08nv40v3t9t9y9[tvtq",
		db: &DB{
			host:         "localhost",
			port:         "5434",
			user:         "postgres",
			password:     "postgres",
			name:         "postgres",
			migrationDir: "file://../../migrations",
		},
	}
}

// DB тест конфиг
type DB struct {
	host         string
	port         string
	user         string
	password     string
	name         string
	connection   *sqlx.DB
	migrationDir string
}

// TestAppConfig -
type TestAppConfig struct {
	logLevel  string
	db        *DB
	jwtSecret string
}

// GetLogLevel возвращает уровень логирования.
func (ac *TestAppConfig) GetLogLevel() string {
	return ac.logLevel
}

// GetHost возвращает хост базы данных.
func (ac *TestAppConfig) GetHost() string {
	return ac.db.host
}

// GetDbPort возвращает порт базы данных.
func (ac *TestAppConfig) GetDbPort() string {
	return ac.db.port
}

// GetUser возвращает имя пользователя базы данных.
func (ac *TestAppConfig) GetUser() string {
	return ac.db.user
}

// GetPassword возвращает пароль базы данных.
func (ac *TestAppConfig) GetPassword() string {
	return ac.db.password
}

// GetDbName возвращает имя базы данных.
func (ac *TestAppConfig) GetDbName() string {
	return ac.db.name
}

// SetDbConnection устанавливает соединение с базой данных.
func (ac *TestAppConfig) SetDbConnection(conn *sqlx.DB) {
	ac.db.connection = conn
}

// GetDbConnection возвращает соединение с базой данных.
func (ac *TestAppConfig) GetDbConnection() *sqlx.DB {
	return ac.db.connection
}

// GetMigrationDir возвращает директорию миграций.
func (ac *TestAppConfig) GetMigrationDir() string {
	return ac.db.migrationDir
}

// GetJWTSecret возвращает секретный ключ JWT.
func (ac *TestAppConfig) GetJWTSecret() string {
	return ac.jwtSecret
}
