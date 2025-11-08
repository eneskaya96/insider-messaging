package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/eneskaya/insider-messaging/pkg/config"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(cfg *config.DatabaseConfig) (*PostgresDB, error) {
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Get().Info("connected to PostgreSQL database",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Name),
	)

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

func (p *PostgresDB) Close() error {
	if p.db != nil {
		logger.Get().Info("closing database connection")
		return p.db.Close()
	}
	return nil
}

func (p *PostgresDB) HealthCheck(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
