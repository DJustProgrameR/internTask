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

func TestOpenReception_Execute(t *testing.T) {
	validPVZID := uuid.New()
	validReceptionID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockReceptionRepo, *mockPVZRepo, *mockTimeService, *mockLogger)
		pvzID         uuid.UUID
		userRole      string
		expected      *onlymodels.Reception
		expectedError string
	}{
		{
			name: "Success - open reception",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("CheckIfExists", mock.Anything, mock.Anything).Return(false, nil)
				mt.On("GetTime").Return(testTime)
				mr.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			pvzID:    validPVZID,
			userRole: model.RoleEmployee.Get(),
			expected: &onlymodels.Reception{
				Id:       &validReceptionID,
				PvzId:    validPVZID,
				DateTime: testTime,
				Status:   onlymodels.InProgress,
			},
		},
		{
			name: "Invalid PVZ ID",
			setupMocks: func(_ *mockReceptionRepo, _ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         uuid.Nil,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid user role",
			setupMocks: func(_ *mockReceptionRepo, _ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Non-employee role",
			setupMocks: func(_ *mockReceptionRepo, _ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "PVZ not found",
			setupMocks: func(_ *mockReceptionRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
				mt.On("GetTime", mock.Anything, mock.Anything).Return(testTime)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInvalidPVZID,
		},
		{
			name: "Error checking PVZ existence",
			setupMocks: func(_ *mockReceptionRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
				mt.On("GetTime", mock.Anything, mock.Anything).Return(testTime)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "Reception already exists",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("CheckIfExists", mock.Anything, mock.Anything).Return(true, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
				mt.On("GetTime", mock.Anything, mock.Anything).Return(testTime)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrReceptionAlreadyOpened,
		},
		{
			name: "Error checking reception existence",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("CheckIfExists", mock.Anything, mock.Anything).Return(false, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
				mt.On("GetTime", mock.Anything, mock.Anything).Return(testTime)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "Error creating reception",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("CheckIfExists", mock.Anything, mock.Anything).Return(false, nil)
				mt.On("GetTime").Return(testTime)
				mr.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &mockReceptionRepo{}
			mz := &mockPVZRepo{}
			mt := &mockTimeService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mr, mz, mt, ml)
			}

			uc := NewUseCaseOpenReception(mr, mz, mt, ml)
			result, err := uc.Execute(context.Background(), tt.pvzID, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.PvzId, result.PvzId)
			assert.Equal(t, tt.expected.DateTime, result.DateTime)
			assert.Equal(t, tt.expected.Status, result.Status)
		})
	}
}

func TestOpenReception_validateInput(t *testing.T) {
	validPVZID := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*mockLogger)
		pvzID         uuid.UUID
		userRole      string
		expectedError string
	}{
		{
			name:     "Valid input",
			pvzID:    validPVZID,
			userRole: model.RoleEmployee.Get(),
		},
		{
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			name:          "Invalid PVZ ID",
			pvzID:         uuid.Nil,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			name:          "Invalid user role",
			pvzID:         validPVZID,
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(ml)
			}
			uc := &OpenReception{logger: ml}

			_, _, err := uc.validateInput(tt.pvzID, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}
