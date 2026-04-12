package app

import (
	"auth-service/internal/migrations"
	"auth-service/internal/repository"
	"auth-service/internal/service"
	"auth-service/internal/transport/grpc"
	api "auth-service/pkg/api/auth/v1"
	"auth-service/pkg/closer"
	"auth-service/pkg/config"
	"auth-service/pkg/logger"
	"auth-service/pkg/migrator"
	"auth-service/pkg/postgres"
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	grpcPort   string
	logs       *slog.Logger
	closer     *closer.Closer
	pool       *pgxpool.Pool
	grpcServer *googleGrpc.Server
}

func NewApp(ctx context.Context) (*App, error) {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		return nil, fmt.Errorf("app.NewApp load config: %w", err)
	}

	logger.Setup(cfg.AppEnv)
	logs := logger.With("service", "auth-service")
	logs.Info("initializing layers",
		"env", cfg.AppEnv,
		"grpc port", cfg.GRPCPort)

	ctx = logger.WithContext(ctx, logs)

	pool, err := postgres.NewPool(ctx, cfg.PGDSN)
	if err != nil {
		return nil, fmt.Errorf("app.NewApp create pool: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()
	m, err := migrator.EmbedMigrations(sqlDB, migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("app.NewApp migrate: %w", err)
	}
	if err := m.Up(); err != nil {
		return nil, fmt.Errorf("app.NewApp migrate: %w", err)
	}

	repo := repository.NewAuthRepo(pool)

	svc := service.NewAuthService(repo, []byte(cfg.JWTSecret))

	handler := grpc.NewAuthHandler(svc)

	server := googleGrpc.NewServer(googleGrpc.UnaryInterceptor(grpc.LoggingInterceptor(logs)))

	api.RegisterAuthServiceServer(server, handler)
	reflection.Register(server)

	cl := closer.New()

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing database connection")
		pool.Close()
		return nil
	})

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing grpc server")
		server.GracefulStop()
		return nil
	})

	return &App{
		grpcPort:   cfg.GRPCPort,
		logs:       logs,
		closer:     cl,
		pool:       pool,
		grpcServer: server,
	}, nil
}

func (a *App) Run() error {
	errCh := make(chan error)

	go func() {
		lis, err := net.Listen("tcp", ":"+a.grpcPort)
		if err != nil {
			errCh <- fmt.Errorf("app.Run net.Listen: %w", err)
			return
		}
		a.logs.Info("starting grpc server",
			slog.String("port", lis.Addr().String()))
		if err := a.grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("app.Run grpcServer.Serve: %w", err)
		}
	}()

	a.logs.Info("app started",
		"port", a.grpcPort)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		a.logs.Error("app.Run server startup failed",
			slog.String("error", err.Error()))
	case sig := <-quit:
		a.logs.Error("app.Run server shutdown",
			slog.Any("signal", sig))
	}

	a.logs.Info("shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.closer.Close(shutdownCtx); err != nil {
		a.logs.Error("shutdown errors", "err", err)
	}

	fmt.Println("Server Stopped")

	return nil
}
