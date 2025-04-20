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

func TestCloseReception_Execute(t *testing.T) {
	validPVZID := uuid.New()
	validReceptionID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockReceptionRepo, *mockPVZRepo, *mockLogger)
		pvzID         uuid.UUID
		userRole      string
		expected      *onlymodels.Reception
		expectedError string
	}{
		{
			name: "Successfully close reception",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
					rec.PVZID = validPVZID.String()
					rec.DateTime = testTime
				}).Return(nil)
				mr.On("Close", mock.Anything, mock.Anything).Return(nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			pvzID:    validPVZID,
			userRole: model.RoleEmployee.Get(),
			expected: &onlymodels.Reception{
				Id:       &validReceptionID,
				PvzId:    validPVZID,
				DateTime: testTime,
				Status:   onlymodels.Close,
			},
		},
		{
			name: "Invalid PVZ ID",
			setupMocks: func(_ *mockReceptionRepo, _ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         uuid.Nil,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid user role",
			setupMocks: func(_ *mockReceptionRepo, _ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Non-employee role",
			setupMocks: func(_ *mockReceptionRepo, _ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "PVZ not found",
			setupMocks: func(_ *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInvalidPVZID,
		},
		{
			name: "Error checking PVZ existence",
			setupMocks: func(_ *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "No active reception found",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Return(nil)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrNoActiveReception,
		},
		{
			name: "Error finding opened reception",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "Error closing reception",
			setupMocks: func(mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mr.On("Close", mock.Anything, mock.Anything).Return(errors.New("db error"))
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
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mr, mz, ml)
			}

			uc := NewUseCaseCloseReception(mr, mz, ml)
			result, err := uc.Execute(context.Background(), tt.pvzID, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.PvzId, result.PvzId)
			assert.Equal(t, tt.expected.Status, result.Status)
		})
	}
}

func TestCloseReception_validateInput(t *testing.T) {
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
			name: "Invalid PVZ ID",
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         uuid.Nil,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid user role",
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
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
			uc := &CloseReception{logger: ml}

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
