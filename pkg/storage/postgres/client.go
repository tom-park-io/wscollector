package postgres

import (
	"context"
	"fmt"

	"wscollector/config"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresClient struct {
	DB *gorm.DB
}

func NewClient(dsn string) (*PostgresClient, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	return &PostgresClient{DB: db}, nil
}

// InitializeAndMigrateKlineRecord connects to Postgres, optionally creates the DB, and runs AutoMigrate.
func InitializeAndMigrateKlineRecord(cfg config.PostgresConfig, createDB bool) (*PostgresClient, error) {
	if createDB {
		if err := CreateDatabase(cfg); err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	}

	client, err := NewClient(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	if err := client.AutoMigrateKlineRecord(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return client, nil
}

func (p *PostgresClient) AutoMigrateKlineRecord() error {
	if err := p.DB.AutoMigrate(&KlineRecord{}); err != nil {
		return fmt.Errorf("auto-migrate kline table: %w", err)
	}
	return nil
}

func (p *PostgresClient) IsHealthy(ctx context.Context) bool {
	db, err := p.DB.DB()
	if err != nil {
		return false
	}
	return db.PingContext(ctx) == nil
}

func (p *PostgresClient) Close() error {
	db, err := p.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to retrieve raw DB: %w", err)
	}
	return db.Close()
}
