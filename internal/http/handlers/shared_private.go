// Package handlers это http хэндлеры
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"time"
)

// Logger логгер
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

const (
	cancelContextTime time.Duration = time.Millisecond * 5000
)

func getUserRoleFromContext(ctx *fiber.Ctx) string {
	userRoleStr := ctx.Locals("userRole").(string)
	return userRoleStr
}
