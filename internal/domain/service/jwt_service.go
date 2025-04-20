// Package service это вспомогательные сервисы
package service

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"time"
)

// CustomClaims данные из токена
type CustomClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService для создания и валидации токенов
type JWTService struct {
	secretKey string
}

// JWTConfig интерфейс конфигурации JWT
type JWTConfig interface {
	GetJWTSecret() string
}

// NewJWTService конструктор для создания нового экземпляра JWTService
func NewJWTService(jwtConfig JWTConfig) JWTService {
	secretKey := jwtConfig.GetJWTSecret()
	if secretKey == "" {
		log.Fatalf("JWTService initialization failed: secret key is empty")
	}
	return JWTService{secretKey: secretKey}
}

// GenerateToken генерирует JWT токен с указанной ролью
func (s JWTService) GenerateToken(role string) (string, error) {
	claims := &CustomClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}

// ValidateToken валидирует JWT токен
func (s JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.secretKey), nil
	})
}

// GetClaims извлекает данные из JWT токена
func (s JWTService) GetClaims(tokenString string) (string, error) {
	token, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims.Role, nil
	}
	return "", err
}
