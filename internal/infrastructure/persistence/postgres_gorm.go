package persistence

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eneskaya/insider-messaging/pkg/config"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type PostgresGormDB struct {
	db *gorm.DB
}

func NewPostgresGormDB(cfg *config.DatabaseConfig) (*PostgresGormDB, error) {
	gormConfig := &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  gormlogger.Warn,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Get().Info("connected to PostgreSQL database with GORM",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.Name),
	)

	return &PostgresGormDB{db: db}, nil
}

func (p *PostgresGormDB) DB() *gorm.DB {
	return p.db
}

func (p *PostgresGormDB) Close() error {
	if p.db != nil {
		sqlDB, err := p.db.DB()
		if err != nil {
			return err
		}
		logger.Get().Info("closing database connection")
		return sqlDB.Close()
	}
	return nil
}

func (p *PostgresGormDB) HealthCheck(ctx context.Context) error {
	sqlDB, err := p.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
