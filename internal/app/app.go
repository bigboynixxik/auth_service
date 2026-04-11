package app

import (
	"auth-service/internal/migrations"
	"auth-service/pkg/closer"
	"auth-service/pkg/config"
	"auth-service/pkg/logger"
	"auth-service/pkg/migrator"
	"auth-service/pkg/postgres"
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	googleGrpc "google.golang.org/grpc"
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

	cl := closer.New()

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing database connection")
		pool.Close()
		return nil
	})

	return &App{
		grpcPort: cfg.GRPCPort,
		logs:     logs,
		closer:   cl,
		pool:     pool,
	}, nil
}

func (a *App) Run() error {

}
