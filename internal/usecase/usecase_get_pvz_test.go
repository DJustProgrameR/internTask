// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	pb "internshipPVZ/internal/grpc/models"
	"internshipPVZ/internal/http/onlymodels"
)

func TestGetPvz_GetFiltered(t *testing.T) {
	validPVZID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockPVZRepo, *mockLogger)
		startDate     string
		endDate       string
		page          int
		limit         int
		userRole      string
		expected      *onlymodels.GetFilteredResponse
		expectedError bool
	}{
		{
			name: "Success - employee role with valid dates",
			setupMocks: func(mz *mockPVZRepo, ml *mockLogger) {
				mz.On("GetAllWithFilter", mock.Anything, "2023-01-01 00:00:00", "2023-01-31 23:59:59", 1, 10).Return([]*dao.PVZList{
					{
						PvzID:             validPVZID.String(),
						RegistrationDate:  testTime,
						City:              model.CityMoscow.ToInt(),
						ReceptionID:       validPVZID.String(),
						ReceptionDateTime: testTime,
						Status:            int8(0),
					},
				}, nil)
				ml.On("Info", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			startDate: "2023-01-01",
			endDate:   "2023-01-31",
			page:      1,
			limit:     10,
			userRole:  model.RoleEmployee.Get(),
			expected: &onlymodels.GetFilteredResponse{
				{
					Pvz: &onlymodels.PVZ{
						Id:               &validPVZID,
						RegistrationDate: &testTime,
						City:             onlymodels.PVZCity(model.CityMoscow.Get()),
					},
				},
			},
		},
		{
			name: "Success - moderator role with no dates",
			setupMocks: func(mz *mockPVZRepo, ml *mockLogger) {
				mz.On("GetAllWithFilter", mock.Anything, "", "", 2, 20).Return([]*dao.PVZList{
					{
						PvzID:             validPVZID.String(),
						RegistrationDate:  testTime,
						City:              model.CitySPB.ToInt(),
						ReceptionID:       validPVZID.String(),
						ReceptionDateTime: testTime,
						Status:            int8(0),
					},
				}, nil)
				ml.On("Info", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			page:     2,
			limit:    20,
			userRole: model.RoleModerator.Get(),
			expected: &onlymodels.GetFilteredResponse{
				{
					Pvz: &onlymodels.PVZ{
						Id:               &validPVZID,
						RegistrationDate: &testTime,
						City:             onlymodels.PVZCity(model.CitySPB.Get()),
					},
				},
			},
		},
		{
			name: "Invalid user role",
			setupMocks: func(_ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			userRole:      "invalid-role",
			expectedError: true,
		},
		{
			name: "Invalid start date format",
			setupMocks: func(_ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			startDate:     "invalid-date",
			userRole:      model.RoleEmployee.Get(),
			expectedError: true,
		},
		{
			name: "Invalid end date format",
			setupMocks: func(_ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			endDate:       "invalid-date",
			userRole:      model.RoleEmployee.Get(),
			expectedError: true,
		},
		{
			name: "Database error",
			setupMocks: func(mz *mockPVZRepo, ml *mockLogger) {
				mz.On("GetAllWithFilter", mock.Anything, "", "", 1, 10).Return(nil, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mz := &mockPVZRepo{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mz, ml)
			}

			uc := NewUseCaseGetPvz(mz, ml)
			result := uc.GetFiltered(context.Background(), tt.startDate, tt.endDate, tt.page, tt.limit, tt.userRole)

			if tt.expectedError {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, len(*tt.expected), len(*result))
			if len(*result) > 0 {
				assert.Equal(t, (*tt.expected)[0].Pvz.Id, (*result)[0].Pvz.Id)
				assert.Equal(t, (*tt.expected)[0].Pvz.City, (*result)[0].Pvz.City)
			}
		})
	}
}

func TestGetPvz_Get(t *testing.T) {
	validPVZID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockPVZRepo, *mockLogger)
		expected      *pb.GetPVZListResponse
		expectedError bool
	}{
		{
			name: "Success - get PVZs",
			setupMocks: func(mz *mockPVZRepo, ml *mockLogger) {
				mz.On("Get", mock.Anything).Return([]*dao.PVZ{
					{
						ID:               validPVZID.String(),
						RegistrationDate: testTime,
						City:             model.CityMoscow.ToInt(),
					},
				}, nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			expected: &pb.GetPVZListResponse{
				Pvzs: []*pb.PVZ{
					{
						Id:               validPVZID.String(),
						RegistrationDate: timestamppb.New(testTime),
						City:             model.CityMoscow.Get(),
					},
				},
			},
		},
		{
			name: "Database error",
			setupMocks: func(mz *mockPVZRepo, ml *mockLogger) {
				mz.On("Get", mock.Anything).Return(nil, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mz := &mockPVZRepo{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mz, ml)
			}

			uc := NewUseCaseGetPvz(mz, ml)
			result := uc.Get(context.Background())

			if tt.expectedError {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, len(tt.expected.Pvzs), len(result.Pvzs))
			if len(result.Pvzs) > 0 {
				assert.Equal(t, tt.expected.Pvzs[0].Id, result.Pvzs[0].Id)
				assert.Equal(t, tt.expected.Pvzs[0].City, result.Pvzs[0].City)
			}
		})
	}
}

func TestGetPvz_validateInput(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockLogger)
		page          int
		limit         int
		userRole      string
		expectedPage  int
		expectedLimit int
		expectedRole  model.Role
		expectedError bool
	}{
		{
			name:          "Valid input - employee",
			page:          1,
			limit:         10,
			userRole:      model.RoleEmployee.Get(),
			expectedPage:  1,
			expectedLimit: 10,
			expectedRole:  model.RoleEmployee,
		},
		{
			name:          "Valid input - moderator",
			page:          2,
			limit:         20,
			userRole:      model.RoleModerator.Get(),
			expectedPage:  2,
			expectedLimit: 20,
			expectedRole:  model.RoleModerator,
		},
		{
			name:          "Default page",
			page:          0,
			limit:         10,
			userRole:      model.RoleEmployee.Get(),
			expectedPage:  1,
			expectedLimit: 10,
			expectedRole:  model.RoleEmployee,
		},
		{
			name:          "Limit too small",
			page:          1,
			limit:         0,
			userRole:      model.RoleEmployee.Get(),
			expectedPage:  1,
			expectedLimit: 10,
			expectedRole:  model.RoleEmployee,
		},
		{
			name:          "Limit too large",
			page:          1,
			limit:         50,
			userRole:      model.RoleEmployee.Get(),
			expectedPage:  1,
			expectedLimit: 10,
			expectedRole:  model.RoleEmployee,
		},
		{
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			name:          "Invalid role",
			userRole:      "invalid-role",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(ml)
			}
			uc := &GetPvz{logger: ml}

			page, limit, role, err := uc.validateInput(tt.page, tt.limit, tt.userRole)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPage, page)
			assert.Equal(t, tt.expectedLimit, limit)
			assert.Equal(t, tt.expectedRole, role)
		})
	}
}

func TestGetPvz_validateTime(t *testing.T) {
	tests := []struct {
		setupMocks    func(*mockLogger)
		name          string
		input         string
		expectedError bool
	}{
		{
			name:  "Valid date",
			input: "2023-01-01",
		},
		{
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			name:          "Invalid date",
			input:         "invalid-date",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(ml)
			}
			uc := &GetPvz{logger: ml}

			_, err := uc.validateTime(tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
