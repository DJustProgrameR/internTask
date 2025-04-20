// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
)

func TestDeleteProduct_Execute(t *testing.T) {
	validPVZID := uuid.New()
	validReceptionID := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*mockProductRepo, *mockReceptionRepo, *mockPVZRepo, *mockLogger)
		pvzID         uuid.UUID
		userRole      string
		expectedError string
	}{
		{
			name: "Success - delete product",
			setupMocks: func(mp *mockProductRepo, mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mp.On("CountProducts", mock.Anything, validReceptionID.String()).Return(1, nil)
				mp.On("DeleteLastFromReception", mock.Anything, validReceptionID.String()).Return(nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			pvzID:    validPVZID,
			userRole: model.RoleEmployee.Get(),
		},
		{
			name: "Invalid PVZ ID",
			setupMocks: func(_ *mockProductRepo, _ *mockReceptionRepo, _ *mockPVZRepo, ml *mockLogger) {
				ml.On("Error", mock.Anything, mock.Anything)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         uuid.Nil,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid user role",
			setupMocks: func(_ *mockProductRepo, _ *mockReceptionRepo, _ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Non-employee role",
			setupMocks: func(_ *mockProductRepo, _ *mockReceptionRepo, _ *mockPVZRepo, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "PVZ not found",
			setupMocks: func(_ *mockProductRepo, _ *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInvalidPVZID,
		},
		{
			name: "Error checking PVZ existence",
			setupMocks: func(_ *mockProductRepo, _ *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "No active reception found",
			setupMocks: func(_ *mockProductRepo, mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
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
			setupMocks: func(_ *mockProductRepo, mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "No products left to delete",
			setupMocks: func(mp *mockProductRepo, mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mp.On("CountProducts", mock.Anything, validReceptionID.String()).Return(0, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrNoProductsLeftToDelete,
		},
		{
			name: "Error counting products",
			setupMocks: func(mp *mockProductRepo, mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mp.On("CountProducts", mock.Anything, validReceptionID.String()).Return(0, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
		{
			name: "Error deleting product",
			setupMocks: func(mp *mockProductRepo, mr *mockReceptionRepo, mz *mockPVZRepo, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mp.On("CountProducts", mock.Anything, validReceptionID.String()).Return(1, nil)
				mp.On("DeleteLastFromReception", mock.Anything, validReceptionID.String()).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			pvzID:         validPVZID,
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &mockProductRepo{}
			mr := &mockReceptionRepo{}
			mz := &mockPVZRepo{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mp, mr, mz, ml)
			}

			uc := NewUseCaseDeleteProduct(mp, mr, mz, ml)
			err := uc.Execute(context.Background(), tt.pvzID, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestDeleteProduct_validateInput(t *testing.T) {
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
			uc := &DeleteProduct{logger: ml}

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
