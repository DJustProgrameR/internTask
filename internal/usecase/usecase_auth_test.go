// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	"internshipPVZ/internal/http/onlymodels"
)

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, user *dao.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*dao.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*dao.User), args.Error(1)
}

type mockHashService struct{ mock.Mock }

func (m *mockHashService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockHashService) HashAndComparePassword(password, hash string) bool {
	args := m.Called(password, hash)
	return args.Bool(0)
}

type mockJWTService struct{ mock.Mock }

func (m *mockJWTService) GenerateToken(role string) (string, error) {
	args := m.Called(role)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
	args := m.Called(tokenString)
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func TestAuth_DummyLogin(t *testing.T) {
	validToken := "valid.token.123"

	tests := []struct {
		name          string
		setupMocks    func(*mockJWTService, *mockLogger)
		request       *onlymodels.PostDummyLoginJSONBody
		expectedToken string
		expectedError string
	}{
		{
			name: "Success - employee role",
			setupMocks: func(mj *mockJWTService, ml *mockLogger) {
				mj.On("GenerateToken", model.RoleEmployee.Get()).Return(validToken, nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostDummyLoginJSONBody{
				Role: onlymodels.PostDummyLoginJSONBodyRoleEmployee,
			},
			expectedToken: validToken,
		},
		{
			name: "Success - moderator role",
			setupMocks: func(mj *mockJWTService, ml *mockLogger) {
				mj.On("GenerateToken", model.RoleModerator.Get()).Return(validToken, nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostDummyLoginJSONBody{
				Role: onlymodels.PostDummyLoginJSONBodyRoleModerator,
			},
			expectedToken: validToken,
		},
		{
			name: "Invalid role",
			setupMocks: func(_ *mockJWTService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostDummyLoginJSONBody{
				Role: "invalid-role",
			},
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Token generation error",
			setupMocks: func(mj *mockJWTService, ml *mockLogger) {
				mj.On("GenerateToken", model.RoleEmployee.Get()).Return("", errors.New("jwt error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostDummyLoginJSONBody{
				Role: onlymodels.PostDummyLoginJSONBodyRoleEmployee,
			},
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &mockUserRepo{}
			mh := &mockHashService{}
			mj := &mockJWTService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mj, ml)
			}

			uc := NewUseCaseAuth(mu, mh, mj, ml)
			token, err := uc.DummyLogin(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestAuth_Register(t *testing.T) {
	validEmail := "test@example.com"
	validPassword := "securePassword123"
	validUserID := uuid.New()
	hashedPassword := "hashedPassword123"

	tests := []struct {
		name          string
		setupMocks    func(*mockUserRepo, *mockHashService, *mockLogger)
		request       *onlymodels.PostRegisterJSONBody
		expectedUser  *onlymodels.User
		expectedError string
	}{
		{
			name: "Success - employee registration",
			setupMocks: func(mu *mockUserRepo, mh *mockHashService, ml *mockLogger) {
				mh.On("HashPassword", validPassword).Return(hashedPassword, nil)
				mu.On("Create", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					user := args.Get(1).(*dao.User)
					user.ID = validUserID.String()
				}).Return(nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
				Role:     onlymodels.Employee,
			},
			expectedUser: &onlymodels.User{
				Id:    &validUserID,
				Email: types.Email(validEmail),
				Role:  onlymodels.UserRoleEmployee,
			},
		},
		{
			name: "Invalid email",
			setupMocks: func(_ *mockUserRepo, _ *mockHashService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email("invalid-email"),
				Password: validPassword,
				Role:     onlymodels.Employee,
			},
			expectedError: model.ErrInvalidEmail,
		},
		{
			name: "Invalid password",
			setupMocks: func(_ *mockUserRepo, _ *mockHashService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email(validEmail),
				Password: "short",
				Role:     onlymodels.Employee,
			},
			expectedError: model.ErrInvalidPassword,
		},
		{
			name: "Invalid role",
			setupMocks: func(_ *mockUserRepo, _ *mockHashService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
				Role:     "invalid-role",
			},
			expectedError: model.ErrInvalidRole,
		},
		{
			name: "Password hashing error",
			setupMocks: func(_ *mockUserRepo, mh *mockHashService, ml *mockLogger) {
				mh.On("HashPassword", validPassword).Return("", errors.New("hashing error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
				Role:     onlymodels.Employee,
			},
			expectedError: model.ErrInternal,
		},
		{
			name: "User already exists",
			setupMocks: func(mu *mockUserRepo, mh *mockHashService, ml *mockLogger) {
				mh.On("HashPassword", validPassword).Return(hashedPassword, nil)
				mu.On("Create", mock.Anything, mock.Anything).Return(errors.New(model.ErrUserWithEmailAlreadyExists))
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
				Role:     onlymodels.Employee,
			},
			expectedError: model.ErrUserWithEmailAlreadyExists,
		},
		{
			name: "Database error",
			setupMocks: func(mu *mockUserRepo, mh *mockHashService, ml *mockLogger) {
				mh.On("HashPassword", validPassword).Return(hashedPassword, nil)
				mu.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostRegisterJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
				Role:     onlymodels.Employee,
			},
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &mockUserRepo{}
			mh := &mockHashService{}
			mj := &mockJWTService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mu, mh, ml)
			}

			uc := NewUseCaseAuth(mu, mh, mj, ml)
			user, err := uc.Register(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedUser.Id, user.Id)
			assert.Equal(t, tt.expectedUser.Email, user.Email)
			assert.Equal(t, tt.expectedUser.Role, user.Role)
		})
	}
}

func TestAuth_Login(t *testing.T) {
	validEmail := "test@example.com"
	validPassword := "securePassword123"
	validToken := "valid.token.123"
	validUserID := uuid.New()
	hashedPassword := "hashedPassword123"

	tests := []struct {
		name          string
		setupMocks    func(*mockUserRepo, *mockHashService, *mockJWTService, *mockLogger)
		request       *onlymodels.PostLoginJSONBody
		expectedToken string
		expectedError string
	}{
		{
			name: "Success - valid credentials",
			setupMocks: func(mu *mockUserRepo, mh *mockHashService, mj *mockJWTService, ml *mockLogger) {
				mu.On("FindByEmail", mock.Anything, validEmail).Return(&dao.User{
					ID:       validUserID.String(),
					Email:    validEmail,
					Password: hashedPassword,
					Role:     model.RoleEmployee.ToInt(),
				}, nil)
				mh.On("HashAndComparePassword", validPassword, hashedPassword).Return(true)
				mj.On("GenerateToken", model.RoleEmployee.Get()).Return(validToken, nil)
				ml.On("Info", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
			},
			expectedToken: validToken,
		},
		{
			name: "Invalid email",
			setupMocks: func(_ *mockUserRepo, _ *mockHashService, _ *mockJWTService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email("invalid-email"),
				Password: validPassword,
			},
			expectedError: model.ErrInvalidEmail,
		},
		{
			name: "Invalid password",
			setupMocks: func(_ *mockUserRepo, _ *mockHashService, _ *mockJWTService, ml *mockLogger) {
				ml.On("Warn", mock.Anything, mock.Anything)
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email(validEmail),
				Password: "short",
			},
			expectedError: model.ErrInvalidPassword,
		},
		{
			name: "User not found",
			setupMocks: func(mu *mockUserRepo, _ *mockHashService, _ *mockJWTService, ml *mockLogger) {
				mu.On("FindByEmail", mock.Anything, validEmail).Return(nil, nil)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
			},
			expectedError: model.ErrEmailOrPasswordIsWrong,
		},
		{
			name: "Database error",
			setupMocks: func(mu *mockUserRepo, _ *mockHashService, _ *mockJWTService, ml *mockLogger) {
				mu.On("FindByEmail", mock.Anything, validEmail).Return(nil, errors.New("db error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
			},
			expectedError: model.ErrInternal,
		},
		{
			name: "Wrong password",
			setupMocks: func(mu *mockUserRepo, mh *mockHashService, _ *mockJWTService, ml *mockLogger) {
				mu.On("FindByEmail", mock.Anything, validEmail).Return(&dao.User{
					ID:       validUserID.String(),
					Email:    validEmail,
					Password: hashedPassword,
					Role:     model.RoleEmployee.ToInt(),
				}, nil)
				mh.On("HashAndComparePassword", validPassword, hashedPassword).Return(false)
				ml.On("Warn", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
			},
			expectedError: model.ErrEmailOrPasswordIsWrong,
		},
		{
			name: "Token generation error",
			setupMocks: func(mu *mockUserRepo, mh *mockHashService, mj *mockJWTService, ml *mockLogger) {
				mu.On("FindByEmail", mock.Anything, validEmail).Return(&dao.User{
					ID:       validUserID.String(),
					Email:    validEmail,
					Password: hashedPassword,
					Role:     model.RoleEmployee.ToInt(),
				}, nil)
				mh.On("HashAndComparePassword", validPassword, hashedPassword).Return(true)
				mj.On("GenerateToken", model.RoleEmployee.Get()).Return("", errors.New("jwt error"))
				ml.On("Error", mock.Anything, mock.Anything)
			},
			request: &onlymodels.PostLoginJSONBody{
				Email:    types.Email(validEmail),
				Password: validPassword,
			},
			expectedError: model.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &mockUserRepo{}
			mh := &mockHashService{}
			mj := &mockJWTService{}
			ml := &mockLogger{}

			if tt.setupMocks != nil {
				tt.setupMocks(mu, mh, mj, ml)
			}

			uc := NewUseCaseAuth(mu, mh, mj, ml)
			token, err := uc.Login(context.Background(), tt.request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}
