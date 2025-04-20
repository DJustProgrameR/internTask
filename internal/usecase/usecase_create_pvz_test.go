// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	"internshipPVZ/internal/http/onlymodels"
)

func TestCreatePVZ_Execute(t *testing.T) {
	validPVZID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockPVZRepo, *mockTimeService, *mockLogger)
		request       *onlymodels.PVZ
		userRole      string
		expected      *onlymodels.PVZ
		expectedError string
	}{
		{
			name: "Success - create PVZ with moderator role",
			setupMocks: func(mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mt.On("GetTime").Return(testTime)
				mz.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					pvz := args.Get(1).(*dao.PVZ)
					pvz.ID = validPVZID.String()
				}).Return(nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CityMoscow.Get()),
			},
			userRole: model.RoleModerator.Get(),
			expected: &onlymodels.PVZ{
				Id:               &validPVZID,
				City:             onlymodels.PVZCity(model.CityMoscow.Get()),
				RegistrationDate: &testTime,
			},
		},
		{
			name: "Invalid user role",
			setupMocks: func(_ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CityMoscow.Get()),
			},
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Non-moderator role",
			setupMocks: func(_ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CityMoscow.Get()),
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid city",
			setupMocks: func(_ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: "invalid-city",
			},
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrInvalidCityName,
		},
		{
			name: "Database error",
			setupMocks: func(mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mt.On("GetTime").Return(testTime)
				mz.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CityMoscow.Get()),
			},
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mz := &mockPVZRepo{}
			mt := &mockTimeService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mz, mt, ml)
			}

			uc := NewUseCaseCreatePVZ(mz, mt, ml)
			result, err := uc.Execute(context.Background(), tt.request, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.City, result.City)
			assert.Equal(t, tt.expected.RegistrationDate, result.RegistrationDate)
		})
	}
}

func TestCreatePVZ_validateInput(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockLogger)
		request       *onlymodels.PVZ
		userRole      string
		expectedCity  model.City
		expectedRole  model.Role
		expectedError string
	}{
		{
			name: "Valid input - Moscow",
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CityMoscow.Get()),
			},
			userRole:     model.RoleModerator.Get(),
			expectedCity: model.CityMoscow,
			expectedRole: model.RoleModerator,
		},
		{
			name: "Valid input - SPB",
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CitySPB.Get()),
			},
			userRole:     model.RoleModerator.Get(),
			expectedCity: model.CitySPB,
			expectedRole: model.RoleModerator,
		},
		{
			name: "Invalid user role",
			setupMocks: func(ml *mockLogger) {
				ml.On("Error", mock.Anything, mock.Anything)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: onlymodels.PVZCity(model.CityMoscow.Get()),
			},
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Invalid city",
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PVZ{
				City: "invalid-city",
			},
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrInvalidCityName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(ml)
			}
			uc := &CreatePVZ{logger: ml}

			pvz, role, err := uc.validateInput(tt.request, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCity, pvz.City)
			assert.Equal(t, tt.expectedRole, role)
		})
	}
}

func TestCreatePVZ_validateCity(t *testing.T) {
	tests := []struct {
		name          string
		city          string
		expected      model.City
		expectedError string
	}{
		{
			name:     "Valid city - Moscow",
			city:     model.CityMoscow.Get(),
			expected: model.CityMoscow,
		},
		{
			name:     "Valid city - SPB",
			city:     model.CitySPB.Get(),
			expected: model.CitySPB,
		},
		{
			name:     "Valid city - Kazan",
			city:     model.CityKazan.Get(),
			expected: model.CityKazan,
		},
		{
			name:          "Invalid city",
			city:          "invalid-city",
			expectedError: model.ErrInvalidCityName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}
			uc := &CreatePVZ{logger: ml}

			result, err := uc.validateCity(tt.city)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
