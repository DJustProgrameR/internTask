// Package service это вспомогательные сервисы
package service

import (
	"golang.org/x/crypto/bcrypt"
)

// HashService для паролей
type HashService struct{}

// NewHashService конструктор для создания нового экземпляра HashService.
func NewHashService() HashService {
	return HashService{}
}

// HashPassword хеширует пароль.
// Возвращает хешированный пароль и ошибку, если таковая возникла.
func (hs HashService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// HashAndComparePassword хеширует пароль и сравнивает его с хешированным паролем.
// Возвращает true, если пароли совпадают, и false в противном случае.
func (hs HashService) HashAndComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
