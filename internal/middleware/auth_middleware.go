// Package middleware это мидлвэр для приложения
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/http/onlymodels"
)

// JWTService токены
type JWTService interface {
	GetClaims(tokenString string) (string, error)
}

// Logger логгер
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// AuthMiddleware возвращает хэндлер для авторизации
func AuthMiddleware(jwtService JWTService, logger Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c == nil {
			logger.Error("received nil ctx",
				"middleware", "AuthMiddleware",
				"error")
			return c.Status(fiber.StatusBadRequest).JSON(onlymodels.Error{Message: model.ErrInternal})
		}
		if c.Path() == "/login" || c.Path() == "/dummyLogin" || c.Path() == "/register" {
			return c.Next()
		}
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusForbidden).JSON(onlymodels.Error{
				Message: model.ErrAccessDenied,
			})
		}

		tokenString := extractToken(authHeader)
		UserRole, err := jwtService.GetClaims(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(onlymodels.Error{
				Message: model.ErrAccessDenied,
			})
		}

		c.Locals("userRole", UserRole)

		return c.Next()
	}
}

func extractToken(authHeader string) string {
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token := authHeader[7:]
		if string(token[0]) == "\"" && string(token[len(token)-1]) == "\"" {
			token = token[1 : len(token)-1]
		}
		return token
	}
	return authHeader
}
