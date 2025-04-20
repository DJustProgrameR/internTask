// Package model это доменные сущности и типы
package model

// тексты ошибок
const (
	ErrInvalidRequest             string = "invalid request. Probably missing required fields or using wrong field values"
	ErrInvalidPVZID               string = "missing or invalid PVZ ID"
	ErrNoActiveReception          string = "no active reception"
	ErrReceptionAlreadyOpened     string = "reception already opened"
	ErrEmailOrPasswordIsWrong     string = "email or password is wrong"
	ErrNoProductsLeftToDelete     string = "no products left to delete"
	ErrAccessDenied               string = "access denied"
	ErrInvalidPassword            string = "password should be 8 to 50 characters long"
	ErrInvalidEmail               string = "email should be *@*.*"
	ErrInvalidRole                string = "invalid or missing role"
	ErrInvalidCityName            string = "missing or invalid city name"
	ErrInvalidProductType         string = "missing or invalid product type"
	ErrInternal                   string = "internal server error"
	ErrUserWithEmailAlreadyExists string = "user with email already exists"
)
