// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	"internshipPVZ/internal/http/onlymodels"
)

func TestPvzDaoToDto(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		testID := uuid.New()
		testTime := time.Now()
		pvzDao := &dao.PVZ{
			ID:               testID.String(),
			RegistrationDate: testTime,
			City:             model.CityMoscow.ToInt(),
		}

		dto, err := pvzDaoToDto(pvzDao)

		assert.NoError(t, err)
		assert.Equal(t, testID, *dto.Id)
		assert.Equal(t, testTime, *dto.RegistrationDate)
		assert.Equal(t, onlymodels.PVZCity("Москва"), dto.City)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		pvzDao := &dao.PVZ{ID: "invalid-uuid"}
		_, err := pvzDaoToDto(pvzDao)
		assert.Error(t, err)
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := pvzDaoToDto(nil)
		assert.Error(t, err)
	})
}

func TestPvzsToGrpcDto(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		testTime := time.Now()
		pvzDaos := []*dao.PVZ{
			{
				ID:               uuid.New().String(),
				RegistrationDate: testTime,
				City:             model.CitySPB.ToInt(),
			},
		}

		dtos, err := pvzsToGrpcDto(pvzDaos)

		assert.NoError(t, err)
		assert.Len(t, dtos, 1)
		assert.Equal(t, pvzDaos[0].ID, dtos[0].Id)
		assert.Equal(t, "Санкт-Петербург", dtos[0].City)
		assert.True(t, testTime.Equal(dtos[0].RegistrationDate.AsTime()))
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := pvzsToGrpcDto(nil)
		assert.Error(t, err)
	})

	t.Run("nil item in slice", func(t *testing.T) {
		_, err := pvzsToGrpcDto([]*dao.PVZ{nil})
		assert.Error(t, err)
	})
}

func TestPvzListToDto(t *testing.T) {
	t.Run("successful conversion with products", func(t *testing.T) {
		pvzID := uuid.New().String()
		recID := uuid.New().String()
		productID := uuid.New().String()
		testTime := time.Now()

		input := []*dao.PVZList{
			{
				PvzID:             pvzID,
				RegistrationDate:  testTime,
				City:              model.CityKazan.ToInt(),
				ReceptionID:       recID,
				ReceptionDateTime: testTime,
				Status:            model.ReceptionInProgress.ToInt(),
				ProductID:         sql.NullString{String: productID, Valid: true},
				ProductDateTime:   sql.NullTime{Time: testTime, Valid: true},
				Type:              sql.NullInt16{Int16: model.ProductShoes.ToInt(), Valid: true},
			},
		}

		result, err := pvzListToDto(input)

		assert.NoError(t, err)
		assert.Len(t, *result, 1)
		assert.Equal(t, "Казань", string((*result)[0].Pvz.City))
		assert.Len(t, *(*result)[0].Receptions, 1)
		assert.NotNil(t, (*(*result)[0].Receptions)[0].Products)
		assert.Equal(t, "обувь", string((*(*(*result)[0].Receptions)[0].Products)[0].Type))
	})

	t.Run("nil input", func(t *testing.T) {
		_, err := pvzListToDto(nil)
		assert.Error(t, err)
	})

	t.Run("nil item in slice", func(t *testing.T) {
		_, err := pvzListToDto([]*dao.PVZList{nil})
		assert.Error(t, err)
	})
}

func TestReceptionDaoToDto(t *testing.T) {
	t.Run("successful conversion", func(t *testing.T) {
		recID := uuid.New().String()
		pvzID := uuid.New().String()
		testTime := time.Now()

		receptionDao := &dao.Reception{
			ID:       recID,
			PVZID:    pvzID,
			DateTime: testTime,
			Status:   model.ReceptionInProgress.ToInt(),
		}

		dto, err := receptionDaoToDto(receptionDao)

		assert.NoError(t, err)
		assert.Equal(t, recID, dto.Id.String())
		assert.Equal(t, pvzID, dto.PvzId.String())
		assert.Equal(t, onlymodels.ReceptionStatus("in_progress"), dto.Status)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		_, err := receptionDaoToDto(&dao.Reception{ID: "invalid"})
		assert.Error(t, err)
	})
}

func TestValidateRawID(t *testing.T) {
	t.Run("valid UUID", func(t *testing.T) {
		testID := uuid.New()
		result, err := validateRawID(testID.String())

		assert.NoError(t, err)
		assert.Equal(t, testID, result)
	})

	t.Run("invalid UUID format", func(t *testing.T) {
		_, err := validateRawID("invalid-uuid")
		assert.Error(t, err)
		assert.Equal(t, model.ErrAccessDenied, err.Error())
	})

	t.Run("empty UUID", func(t *testing.T) {
		_, err := validateRawID("")
		assert.Error(t, err)
	})
}

func TestValidateID(t *testing.T) {
	t.Run("valid v4 UUID", func(t *testing.T) {
		testID := uuid.New()
		result, err := validateID(testID)

		assert.NoError(t, err)
		assert.Equal(t, testID, result)
	})

	t.Run("nil UUID", func(t *testing.T) {
		_, err := validateID(uuid.Nil)
		assert.Error(t, err)
	})

	t.Run("invalid UUID version", func(t *testing.T) {
		// Другая версия uuid
		v1UUID := uuid.Must(uuid.NewUUID())
		_, err := validateID(v1UUID)
		assert.Error(t, err)
	})
}

func TestValidateRole(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected model.Role
		err      error
	}{
		{"employee role", "employee", model.RoleEmployee, nil},
		{"moderator role", "moderator", model.RoleModerator, nil},
		{"invalid role", "admin", model.RoleDefault, errors.New(model.ErrInvalidRole)},
		{"empty role", "", model.RoleDefault, errors.New(model.ErrInvalidRole)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateRole(tt.input)

			if tt.err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.err.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}
