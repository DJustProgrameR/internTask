// Package model это доменные сущности и типы
package model

import (
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/repository/dao"
)

// Role виды ролей
type Role string

// роли сотрудников
const (
	RoleEmployee  Role = "employee"
	RoleModerator Role = "moderator"
	RoleDefault   Role = ""
)

// Get возвращает строковое представление роли.
func (r Role) Get() string {
	return string(r)
}

// ToInt возвращает целочисленное представление роли.
func (r Role) ToInt() int8 {
	mapp := map[Role]int8{
		RoleEmployee:  0,
		RoleModerator: 1,
		RoleDefault:   2,
	}
	return mapp[r]
}

// NewUserRole конструктор для создания роли пользователя из целочисленного значения.
func NewUserRole(num int8) Role {
	mapp := map[int8]Role{
		0: RoleEmployee,
		1: RoleModerator,
		2: RoleDefault,
	}
	return mapp[num]
}

// User сущность пользователя
type User struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     Role
}

// ToDao преобразует сущность пользователя в DAO объект.
func (u User) ToDao() *dao.User {
	return &dao.User{ID: u.ID.String(), Email: u.Email, Role: u.Role.ToInt(), Password: u.Password}
}
