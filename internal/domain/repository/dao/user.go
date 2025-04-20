// Package dao это dao для общения с репозиториями
package dao

// User dao
type User struct {
	ID       string
	Email    string
	Password string
	Role     int8
}
