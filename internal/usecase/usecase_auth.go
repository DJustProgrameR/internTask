// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oapi-codegen/runtime/types"
	"internshipPVZ/internal/domain/model"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/domain/repository/dao"
	"internshipPVZ/internal/http/onlymodels"
	"log"
	"regexp"
)

// JWTService для генерации и проверки токенов
type JWTService interface {
	GenerateToken(role string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

// HashService для хэширования пароля
type HashService interface {
	HashPassword(password string) (string, error)
	HashAndComparePassword(password, hash string) bool
}

// Auth юзкейс
type Auth struct {
	userRepo    repo.UserRepo
	hashService HashService
	authService JWTService
	emailRegex  *regexp.Regexp
	logger      Logger
}

// NewUseCaseAuth конструктор
func NewUseCaseAuth(
	userRepo repo.UserRepo,
	hashService HashService,
	authService JWTService,
	logger Logger,
) *Auth {
	if userRepo == nil {
		log.Fatalf("Auth usecase userRepo nil")

	}
	if hashService == nil {
		log.Fatalf("Auth usecase hashService nil")

	}
	if authService == nil {
		log.Fatalf("Auth usecase authService nil")

	}
	if logger == nil {
		log.Fatalf("Auth usecase logger nil")

	}

	return &Auth{
		userRepo:    userRepo,
		hashService: hashService,
		authService: authService,
		emailRegex:  regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		logger:      logger,
	}
}

// DummyLogin тупо выдаёт токен
func (uc *Auth) DummyLogin(_ context.Context, request *onlymodels.PostDummyLoginJSONBody) (string, error) {
	user, err := uc.validateDummyLoginInput(request)
	if err != nil {
		uc.logger.Error("dummy login validation failed",
			"usecase", "Auth",
			"method", "validateDummyLoginInput",
			"error", err)
		return "", err
	}

	token, err := uc.authService.GenerateToken(user.Role.Get())
	if err != nil {
		uc.logger.Error("failed to generate token",
			"usecase", "Auth",
			"method", "authService.GenerateToken",
			"error", err)
		return "", errors.New(model.ErrInternal)
	}

	uc.logger.Info("dummy login successful",
		"usecase", "Auth",
		"role", user.Role.Get())
	return token, nil
}

// Register регистрирует пользователя
func (uc *Auth) Register(ctx context.Context, request *onlymodels.PostRegisterJSONBody) (*onlymodels.User, error) {
	user, err := uc.validateRegisterInput(request)
	if err != nil {
		uc.logger.Error("register validation failed",
			"usecase", "Auth",
			"method", "validateRegisterInput",
			"error", err)
		return nil, err
	}

	user.Password, err = uc.hashService.HashPassword(user.Password)
	if err != nil {
		uc.logger.Error("failed to hash password",
			"usecase", "Auth",
			"method", "hashService.HashPassword",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	userDao := user.ToDao()

	err = uc.userRepo.Create(ctx, userDao)
	if err != nil {
		if err.Error() == model.ErrUserWithEmailAlreadyExists {
			uc.logger.Warn("user already exists",
				"usecase", "Auth",
				"method", "userRepo.Create",
				"email", user.Email)
			return nil, err
		}
		uc.logger.Error("failed to create user",
			"usecase", "Auth",
			"method", "userRepo.Create",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	userDto, err := uc.userDaoToDto(userDao)
	if err != nil {
		uc.logger.Error("failed to convert user DAO to DTO",
			"usecase", "Auth",
			"method", "userDaoToDto",
			"error", err)
		return nil, errors.New(model.ErrInternal)
	}

	uc.logger.Info("user registered successfully",
		"usecase", "Auth",
		"user_id", userDao.ID,
		"email", user.Email,
		"role", user.Role.Get())
	return userDto, nil
}

// Login выдаёт соотв токен по данным пользователя
func (uc *Auth) Login(ctx context.Context, request *onlymodels.PostLoginJSONBody) (string, error) {
	user, err := uc.validateLoginInput(request)
	if err != nil {
		uc.logger.Error("login validation failed",
			"usecase", "Auth",
			"method", "validateLoginInput",
			"error", err)
		return "", err
	}

	foundUser, err := uc.userRepo.FindByEmail(ctx, user.Email)
	if err != nil {
		uc.logger.Error("failed to find user by email",
			"usecase", "Auth",
			"method", "userRepo.FindByEmail",
			"email", user.Email,
			"error", err)
		return "", errors.New(model.ErrInternal)
	}

	if foundUser == nil {
		uc.logger.Warn("user not found",
			"usecase", "Auth",
			"method", "Login",
			"email", user.Email)
		return "", errors.New(model.ErrEmailOrPasswordIsWrong)
	}

	auth := uc.hashService.HashAndComparePassword(user.Password, foundUser.Password)
	if !auth {
		uc.logger.Warn("invalid password",
			"usecase", "Auth",
			"method", "Login",
			"email", user.Email)
		return "", errors.New(model.ErrEmailOrPasswordIsWrong)
	}

	token, err := uc.authService.GenerateToken(model.NewUserRole(foundUser.Role).Get())
	if err != nil {
		uc.logger.Error("failed to generate token",
			"usecase", "Auth",
			"method", "authService.GenerateToken",
			"error", err)
		return "", errors.New(model.ErrInternal)
	}

	uc.logger.Info("user logged in successfully",
		"usecase", "Auth",
		"user_id", foundUser.ID,
		"email", user.Email)
	return token, nil
}

func (uc *Auth) userDaoToDto(user *dao.User) (*onlymodels.User, error) {
	id, err := validateRawID(user.ID)
	if err != nil {
		uc.logger.Error("failed to parse user ID",
			"usecase", "Auth",
			"method", "userDaoToDto",
			"user_id", user.ID,
			"error", err)
		return nil, err
	}
	return &onlymodels.User{
		Id:    &id,
		Email: types.Email(user.Email),
		Role:  onlymodels.UserRole(model.NewUserRole(user.Role)),
	}, nil
}

func (uc *Auth) validateDummyLoginInput(request *onlymodels.PostDummyLoginJSONBody) (user *model.User, err error) {
	user = &model.User{}
	if user.Role, err = validateRole(string(request.Role)); err != nil {
		uc.logger.Warn("invalid role in dummy login",
			"usecase", "Auth",
			"method", "validateDummyLoginInput",
			"requested_role", request.Role)
		return
	}
	return
}

func (uc *Auth) validateRegisterInput(request *onlymodels.PostRegisterJSONBody) (user model.User, err error) {
	user = model.User{}
	if user.Email, err = uc.validateEmail(string(request.Email)); err != nil {
		uc.logger.Warn("invalid email in registration",
			"usecase", "Auth",
			"method", "validateRegisterInput",
			"email", request.Email)
		return
	}
	if user.Password, err = uc.validatePassword(request.Password); err != nil {
		uc.logger.Warn("invalid password in registration",
			"usecase", "Auth",
			"method", "validateRegisterInput")
		return
	}
	if user.Role, err = validateRole(string(request.Role)); err != nil {
		uc.logger.Warn("invalid role in registration",
			"usecase", "Auth",
			"method", "validateRegisterInput",
			"requested_role", request.Role)
		return
	}
	return
}

func (uc *Auth) validateLoginInput(request *onlymodels.PostLoginJSONBody) (user *model.User, err error) {
	user = &model.User{}
	if user.Email, err = uc.validateEmail(string(request.Email)); err != nil {
		uc.logger.Warn("invalid email in login",
			"usecase", "Auth",
			"method", "validateLoginInput",
			"email", request.Email)
		return
	}
	if user.Password, err = uc.validatePassword(request.Password); err != nil {
		uc.logger.Warn("invalid password in login",
			"usecase", "Auth",
			"method", "validateLoginInput")
		return
	}
	return
}

func (uc *Auth) validateEmail(email string) (string, error) {
	if !uc.emailRegex.MatchString(email) {
		return "", errors.New(model.ErrInvalidEmail)
	}
	return email, nil
}

func (uc *Auth) validatePassword(password string) (string, error) {
	if len(password) < 8 || len(password) > 50 {
		return "", errors.New(model.ErrInvalidPassword)
	}
	return password, nil
}
