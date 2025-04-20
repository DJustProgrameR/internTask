// Package service это вспомогательные сервисы
package service

import (
	"log/slog"
	"os"
)

// LoggerConfig интерфейс конфигурации логгера
type LoggerConfig interface {
	GetLogLevel() string
}

// newLogger создает новый логгер на основе конфигурации
func newLogger(config LoggerConfig) *slog.Logger {
	var log *slog.Logger

	switch config.GetLogLevel() {
	case "local":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	case "dev":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}),
		)
	case "prod":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelInfo,
			}),
		)
	}

	return log
}

// NewSlogLogger конструктор для создания нового экземпляра SlogLogger
func NewSlogLogger(config LoggerConfig) *SlogLogger {
	return &SlogLogger{logger: newLogger(config)}
}

// SlogLogger логгер-обёртка
type SlogLogger struct {
	logger *slog.Logger
}

// Debug логирует сообщение уровня отладки
func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info логирует информационное сообщение
func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn логирует предупреждающее сообщение
func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error логирует сообщение об ошибке
func (l *SlogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}
