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

type mockReceptionRepo struct{ mock.Mock }

func (m *mockReceptionRepo) Create(ctx context.Context, reception *dao.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *mockReceptionRepo) FindOpened(ctx context.Context, reception *dao.Reception) error {
	args := m.Called(ctx, reception)
	return args.Error(0)
}

func (m *mockReceptionRepo) CheckIfExists(ctx context.Context, reception *dao.Reception) (bool, error) {
	args := m.Called(ctx, reception)
	return args.Bool(0), args.Error(1)
}

func (m *mockReceptionRepo) Close(ctx context.Context, rec *dao.Reception) error {
	args := m.Called(ctx, rec)
	return args.Error(0)
}

type mockProductRepo struct{ mock.Mock }

func (m *mockProductRepo) Add(ctx context.Context, product *dao.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *mockProductRepo) CountProducts(ctx context.Context, receptionID string) (int, error) {
	args := m.Called(ctx, receptionID)
	return args.Int(0), args.Error(1)
}

func (m *mockProductRepo) DeleteLastFromReception(ctx context.Context, receptionID string) error {
	args := m.Called(ctx, receptionID)
	return args.Error(0)
}

type mockPVZRepo struct{ mock.Mock }

func (m *mockPVZRepo) Create(ctx context.Context, pvz *dao.PVZ) error {
	args := m.Called(ctx, pvz)
	return args.Error(0)
}

func (m *mockPVZRepo) GetAllWithFilter(ctx context.Context, startDate, endDate string, page, limit int) ([]*dao.PVZList, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*dao.PVZList), args.Error(1)
}

func (m *mockPVZRepo) Get(ctx context.Context) ([]*dao.PVZ, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dao.PVZ), args.Error(1)
}

func (m *mockPVZRepo) CheckIfExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type mockTimeService struct{ mock.Mock }

func (m *mockTimeService) GetTime() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

type mockLogger struct{ mock.Mock }

func (m *mockLogger) Debug(msg string, args ...any) { m.Called(msg, args) }
func (m *mockLogger) Info(msg string, args ...any)  { m.Called(msg, args) }
func (m *mockLogger) Warn(msg string, args ...any)  { m.Called(msg, args) }
func (m *mockLogger) Error(msg string, args ...any) { m.Called(msg, args) }

func TestAddProduct_Execute(t *testing.T) {
	validPVZID := uuid.New()
	validReceptionID := uuid.New()
	validProductID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		setupMocks    func(*mockReceptionRepo, *mockProductRepo, *mockPVZRepo, *mockTimeService, *mockLogger)
		request       *onlymodels.PostProductsJSONBody
		userRole      string
		expected      *onlymodels.Product
		expectedError string
	}{
		{
			name: "Successfully add product",
			setupMocks: func(mr *mockReceptionRepo, mp *mockProductRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mp.On("Add", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					p := args.Get(1).(*dao.Product)
					p.ID = validProductID.String()
				}).Return(nil)
				mt.On("GetTime").Return(testTime)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole: model.RoleEmployee.Get(),
			expected: &onlymodels.Product{
				Id:          &validProductID,
				DateTime:    &testTime,
				Type:        "электроника",
				ReceptionId: validReceptionID,
			},
		},
		{
			name: "Invalid PVZ ID",
			setupMocks: func(_ *mockReceptionRepo, _ *mockProductRepo, _ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: uuid.Nil,
				Type:  "электроника",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid product type",
			setupMocks: func(_ *mockReceptionRepo, _ *mockProductRepo, _ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Error", mock.Anything, mock.Anything)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "invalid-type",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInvalidProductType,
		},
		{
			name: "Non-employee role",
			setupMocks: func(_ *mockReceptionRepo, _ *mockProductRepo, _ *mockPVZRepo, _ *mockTimeService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole:      model.RoleModerator.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "PVZ not found",
			setupMocks: func(_ *mockReceptionRepo, _ *mockProductRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(false, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
				mt.On("GetTime").Return(time.Now())
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInvalidPVZID,
		},
		{
			name: "No active reception",
			setupMocks: func(mr *mockReceptionRepo, _ *mockProductRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Return(nil)
				ml.On("Error", mock.Anything, mock.Anything)
				ml.On("Warn", mock.Anything, mock.Anything)
				mt.On("GetTime").Return(time.Now())
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrNoActiveReception,
		},
		{
			name: "Error adding product",
			setupMocks: func(mr *mockReceptionRepo, mp *mockProductRepo, mz *mockPVZRepo, mt *mockTimeService, ml *mockLogger) {
				mz.On("CheckIfExists", mock.Anything, validPVZID.String()).Return(true, nil)
				mr.On("FindOpened", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					rec := args.Get(1).(*dao.Reception)
					rec.ID = validReceptionID.String()
				}).Return(nil)
				mp.On("Add", mock.Anything, mock.Anything).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
				mt.On("GetTime").Return(time.Now())
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := &mockReceptionRepo{}
			mp := &mockProductRepo{}
			mz := &mockPVZRepo{}
			mt := &mockTimeService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mr, mp, mz, mt, ml)
			}

			uc := NewUseCaseAddProduct(mr, mp, mz, mt, ml)
			result, err := uc.Execute(context.Background(), tt.request, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.DateTime, result.DateTime)
			assert.Equal(t, tt.expected.ReceptionId, result.ReceptionId)
		})
	}
}

func TestAddProduct_validateInput(t *testing.T) {
	validPVZID := uuid.New()

	tests := []struct {
		name          string
		setupMocks    func(*mockTimeService, *mockLogger)
		request       *onlymodels.PostProductsJSONBody
		userRole      string
		expectedError string
	}{
		{
			name: "Valid input",
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole: model.RoleEmployee.Get(),
		},
		{
			name: "Invalid PVZ ID",
			request: &onlymodels.PostProductsJSONBody{
				PvzId: uuid.Nil,
				Type:  "электроника",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrAccessDenied,
		},
		{
			name: "Invalid user role",
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "электроника",
			},
			userRole:      "invalid-role",
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Invalid product type",
			setupMocks: func(mt *mockTimeService, ml *mockLogger) {
				mt.On("GetTime").Return(time.Now())
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostProductsJSONBody{
				PvzId: validPVZID,
				Type:  "invalid-type",
			},
			userRole:      model.RoleEmployee.Get(),
			expectedError: model.ErrInvalidProductType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &mockTimeService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mt, ml)
			}
			uc := &AddProduct{logger: ml}

			_, _, _, err := uc.validateInput(tt.request, tt.userRole)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestAddProduct_validateProductType(t *testing.T) {
	tests := []struct {
		name          string
		productType   string
		setupMocks    func(*mockLogger)
		expected      model.ProductType
		expectedError string
	}{
		{
			name:        "Valid electronics",
			productType: "электроника",
			expected:    model.ProductElectronics,
		},
		{
			name:        "Valid clothes",
			productType: "одежда",
			expected:    model.ProductClothes,
		},
		{
			name:        "Valid shoes",
			productType: "обувь",
			expected:    model.ProductShoes,
		},
		{
			name: "Invalid type",
			setupMocks: func(ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			productType:   "invalid",
			expectedError: model.ErrInvalidProductType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(ml)
			}
			uc := &AddProduct{logger: ml}

			result, err := uc.validateProductType(tt.productType)

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

func TestAddProduct_productDaoToDto(t *testing.T) {
	validProductID := uuid.New()
	validReceptionID := uuid.New()
	testTime := time.Now()

	tests := []struct {
		name          string
		product       *dao.Product
		expected      *onlymodels.Product
		expectedError string
	}{
		{
			name: "Valid conversion",
			product: &dao.Product{
				ID:          validProductID.String(),
				DateTime:    testTime,
				ReceptionID: validReceptionID.String(),
				Type:        model.ProductElectronics.ToInt(),
			},
			expected: &onlymodels.Product{
				Id:          &validProductID,
				DateTime:    &testTime,
				Type:        "электроника",
				ReceptionId: validReceptionID,
			},
		},
		{
			name:          "Nil product",
			product:       nil,
			expectedError: "product is nil",
		},
		{
			name: "Invalid product ID",
			product: &dao.Product{
				ID:          "invalid",
				DateTime:    testTime,
				ReceptionID: validReceptionID.String(),
				Type:        model.ProductElectronics.ToInt(),
			},
			expectedError: model.ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ml := &mockLogger{}
			uc := &AddProduct{logger: ml}

			result, err := uc.productDaoToDto(tt.product)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Id, result.Id)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.DateTime, result.DateTime)
			assert.Equal(t, tt.expected.ReceptionId, result.ReceptionId)
		})
	}
}
