// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"errors"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
)

func validateRawID(idr string) (uuid.UUID, error) {
	var id uuid.UUID
	var err error
	if id, err = uuid.Parse(idr); err != nil {
		return uuid.New(), errors.New(model.ErrAccessDenied)
	}
	return validateID(id)
}

func validateID(id uuid.UUID) (uuid.UUID, error) {
	if id == uuid.Nil {
		return uuid.New(), errors.New(model.ErrAccessDenied)
	}

	version := id.Version()

	variant := id.Variant()

	if version != 4 || variant != uuid.RFC4122 {
		return uuid.New(), errors.New(model.ErrAccessDenied)
	}
	return id, nil
}

func validateRole(role string) (model.Role, error) {
	switch role {
	case model.RoleEmployee.Get():
		return model.RoleEmployee, nil
	case model.RoleModerator.Get():
		return model.RoleModerator, nil
	default:
		return model.RoleDefault, errors.New(model.ErrInvalidRole)
	}
}
