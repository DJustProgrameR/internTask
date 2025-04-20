// Package integration это интеграционное тестирование
package integration

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"internshipPVZ/cmd/initdb"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/domain/service"
	"internshipPVZ/internal/grpc/handler"
	"internshipPVZ/internal/http"
	"internshipPVZ/internal/http/handlers"
	"internshipPVZ/internal/repository"
	"internshipPVZ/internal/usecase"
	"log"
	"testing"
	"time"
)

// TestAppModule тестовая сборка DI контейнера
type TestAppModule struct {
}

func (am *TestAppModule) Invoke(t *testing.T) {
	app := fx.New(
		ModuleConfig(),
		fx.Invoke(initializeDatabase),
		Module(t),
		fx.Invoke(registerHTTPServer),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start app: %v", err)
	}

	<-ctx.Done()

	if err := app.Stop(context.Background()); err != nil {
		t.Fatalf("Failed to stop app: %v", err)
	}
}

// AppConfig тестовый конфиг приложения
type AppConfig interface {
	GetHost() string
	GetDbPort() string
	GetUser() string
	GetPassword() string
	GetDbName() string
	SetDbConnection(conn *sqlx.DB)
	GetDbConnection() *sqlx.DB
	GetMigrationDir() string
	GetJWTSecret() string
}

func ModuleConfig() fx.Option {
	return fx.Options(
		// Конфиг приложения
		fx.Provide(fx.Annotate(
			NewTestAppConfig,
			fx.As(new(service.JWTConfig)),
			fx.As(new(repository.Config)),
			fx.As(new(AppConfig)),
			fx.As(new(service.LoggerConfig)))),
		// Регистрируем сервисы
		fx.Provide(fx.Annotate(
			service.NewTimeService,
			fx.As(new(usecase.TimeService)))),
		fx.Provide(fx.Annotate(
			service.NewJWTService,
			fx.As(new(usecase.JWTService)),
			fx.As(new(http.JWTService)))),
		fx.Provide(fx.Annotate(
			service.NewHashService,
			fx.As(new(usecase.HashService)))),
		fx.Provide(fx.Annotate(
			service.NewSlogLogger,
			fx.As(new(usecase.Logger)),
			fx.As(new(handlers.Logger)))),
	)
}

func Module(t *testing.T) fx.Option {
	return fx.Options(
		// Регистрируем репозитории
		fx.Provide(fx.Annotate(
			repository.NewProductRepo,
			fx.As(new(repo.ProductRepo)))),
		fx.Provide(fx.Annotate(
			repository.NewPVZRepo,
			fx.As(new(repo.PVZRepo)))),
		fx.Provide(fx.Annotate(
			repository.NewUserRepo,
			fx.As(new(repo.UserRepo)))),
		fx.Provide(fx.Annotate(
			repository.NewReceptionRepo,
			fx.As(new(repo.ReceptionRepo)))),
		// Регистрируем юзкейсы
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseAuth,
			fx.As(new(handlers.AuthUseCase)))),
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseCreatePVZ,
			fx.As(new(handlers.CreatePVZUseCase)))),
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseGetPvz,
			fx.As(new(handlers.GetPVZUseCase)),
			fx.As(new(handler.GetPvzUseCase)))),
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseOpenReception,
			fx.As(new(handlers.OpenReceptionUseCase)))),
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseAddProduct,
			fx.As(new(handlers.AddProductUsecase)))),
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseDeleteProduct,
			fx.As(new(handlers.DeleteProductUseCase)))),
		fx.Provide(fx.Annotate(
			usecase.NewUseCaseCloseReception,
			fx.As(new(handlers.CloseReceptionUseCase)))),
		// Регистрируем http хэндлеры
		fx.Provide(fx.Annotate(
			handlers.NewProductController,
			fx.As(new(http.ProductController)))),
		fx.Provide(fx.Annotate(
			handlers.NewPVZController,
			fx.As(new(http.PVZController)))),
		fx.Provide(fx.Annotate(
			handlers.NewAuthController,
			fx.As(new(http.AuthController)))),
		fx.Provide(fx.Annotate(
			handlers.NewReceptionController,
			fx.As(new(http.ReceptionController)))),
		// Регистрируем тест
		fx.Provide(
			ProvideTest(t)),
		// Регистрируем http сервер приложения
		fx.Provide(fx.Annotate(
			http.NewHTTPServer)),
	)
}

func ProvideTest(t *testing.T) func() *testing.T {
	return func() *testing.T {
		return t
	}
}

func initializeDatabase(cfg AppConfig) {
	if cfg == nil {
		log.Fatalf("initializeDatabase failed: AppConfig is nil")
	}
	dbConn, err := initdb.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}
	cfg.SetDbConnection(dbConn)

	if err := initdb.RunMigrations(cfg); err != nil {
		log.Fatalf(err.Error())
	}
}

func registerHTTPServer(t *testing.T, lc fx.Lifecycle, testApp *fiber.App, cfg AppConfig) {
	if testApp == nil {
		log.Fatalf("registerHTTPServer failed: HttpServer is nil")
	}
	if cfg == nil {
		log.Fatalf("registerHTTPServer failed: AppConfig is nil")
	}
	FullFlowTest(t, testApp)

	if err := testApp.Shutdown(); err != nil {
		t.Errorf("Failed to shutdown Fiber app: %v", err)
	}

	err := initdb.DropDBTables(cfg)
	if err != nil {
		t.Errorf("Failed to close DB connection: %v", err)
	}

	err = cfg.GetDbConnection().Close()

	if err != nil {
		t.Errorf("Failed to close DB connection: %v", err)
	}
}
