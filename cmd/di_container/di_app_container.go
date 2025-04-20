// Package di_container это DI контейнер приложения
package di_container

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"internshipPVZ/cmd/config"
	"internshipPVZ/cmd/initdb"
	repo "internshipPVZ/internal/domain/repository"
	"internshipPVZ/internal/domain/service"
	"internshipPVZ/internal/grpc"
	"internshipPVZ/internal/grpc/handler"
	"internshipPVZ/internal/http"
	"internshipPVZ/internal/http/handlers"
	"internshipPVZ/internal/prometheus"
	"internshipPVZ/internal/repository"
	"internshipPVZ/internal/usecase"
	"log"
)

// AppModule контейнер приложения
type AppModule struct {
}

// Invoke запускает DI контейнер приложения.
func (am *AppModule) Invoke() {
	fx.New(
		ModuleConfig(),
		fx.Invoke(initializeDatabase),
		Module(),
		fx.Invoke(registerHTTPServer),
		fx.Invoke(registerGrpcServer),
		fx.Invoke(registerPrometheusServer),
	).Run()
}

// AppConfig интерфейс конфигурации приложения.
type AppConfig interface {
	GetAppPort() string
	GetHost() string
	GetDbPort() string
	GetGrpcPort() string
	GetPrometheusPort() string
	GetUser() string
	GetPassword() string
	GetDbName() string
	SetDbConnection(conn *sqlx.DB)
	GetDbConnection() *sqlx.DB
	GetMigrationDir() string
	GetJWTSecret() string
}

// HttpServer интерфейс для HTTP сервера.
type HttpServer interface {
	Listen(addr string) error
	Shutdown() error
}

// GrpcServer интерфейс для gRPC сервера.
type GrpcServer interface {
	Listen(portString string) error
	Shutdown() error
}

// MetricsServer интерфейс для сервера метрик.
type MetricsServer interface {
	Listen(addr string) error
	Shutdown() error
}

// ModuleConfig возвращает опции для конфигурации модуля.
func ModuleConfig() fx.Option {
	return fx.Options(
		// Конфиг приложения
		fx.Provide(fx.Annotate(
			config.NewAppConfig,
			fx.As(new(service.JWTConfig)),
			fx.As(new(repository.RepositoryConfig)),
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

// Module возвращает опции для модуля.
func Module() fx.Option {
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
		// Регистрируем HTTP хэндлеры
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
		// Регистрируем HTTP сервер приложения
		fx.Provide(fx.Annotate(
			http.NewHTTPServer,
			fx.As(new(HttpServer)))),
		// Регистрируем Prometheus сервер
		fx.Provide(fx.Annotate(
			prometheus.NewPrometheusServer,
			fx.As(new(MetricsServer)))),
		// Регистрируем gRPC сервер
		fx.Provide(fx.Annotate(
			grpc.NewServer,
			fx.As(new(GrpcServer)))),
	)
}

// initializeDatabase инициализирует соединение с базой данных и выполняет миграции.
func initializeDatabase(cfg AppConfig) {
	if cfg == nil {
		log.Fatalf("initializeDatabase failed: AppConfig is nil")
	}
	dbConn, err := initdb.NewDBConnection(cfg)
	if err != nil {
		log.Println(err)
	}
	cfg.SetDbConnection(dbConn)

	if err := initdb.RunMigrations(cfg); err != nil {
		log.Println(err)
	}
}

// registerHTTPServer регистрирует HTTP сервер в DI контейнере.
func registerHTTPServer(lc fx.Lifecycle, httpServer HttpServer, cfg AppConfig) {
	if httpServer == nil {
		log.Fatalf("registerHTTPServer failed: HttpServer is nil")
	}
	if cfg == nil {
		log.Fatalf("registerHTTPServer failed: AppConfig is nil")
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			httpPort := fmt.Sprintf(":%s", cfg.GetAppPort())
			go func() {
				if err := httpServer.Listen(httpPort); err != nil {
					log.Println("Http server failed:", err)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			err := httpServer.Shutdown()
			if err != nil {
				return err
			}
			err = initdb.DropDBTables(cfg)
			if err != nil {
				return err
			}
			return cfg.GetDbConnection().Close()
		},
	})
}

// registerGrpcServer регистрирует gRPC сервер в DI контейнере.
func registerGrpcServer(lc fx.Lifecycle, grpcServer GrpcServer, cfg AppConfig) {
	if cfg == nil {
		log.Fatalf("registerGrpcServer failed: AppConfig is nil")
	}
	if grpcServer == nil {
		log.Fatalf("registerGrpcServer failed: GrpcServer is nil")
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				addr := fmt.Sprintf(":%s", cfg.GetGrpcPort())
				log.Println("[Grpc] Serving at", addr)
				if err := grpcServer.Listen(addr); err != nil {
					log.Println("Grpc server failed:", err)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			return grpcServer.Shutdown()
		},
	})
}

// registerPrometheusServer регистрирует сервер метрик в DI контейнере.
func registerPrometheusServer(lc fx.Lifecycle, metricsServer MetricsServer, cfg AppConfig) {
	if metricsServer == nil {
		log.Fatalf("registerPrometheusServer failed: MetricsServer is nil")
	}
	if cfg == nil {
		log.Fatalf("registerPrometheusServer failed: AppConfig is nil")
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				addr := fmt.Sprintf(":%s", cfg.GetPrometheusPort())
				log.Println("[Prometheus] Serving metrics at", addr)
				if err := metricsServer.Listen(addr); err != nil {
					log.Println("Prometheus server failed:", err)
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			return metricsServer.Shutdown()
		},
	})
}
