package repository

import (
	"context"
	"sync"

	"github.com/gofreego/goutils/databases"
	sqlutils "github.com/gofreego/goutils/databases/connections/sql"
	"github.com/gofreego/openclick/internal/repository/clickhouse"
	"github.com/gofreego/openclick/internal/repository/postgresql"
	"github.com/gofreego/openclick/internal/service"
)

// Config holds repository selection and connection config
type Config struct {
	Name       string            `yaml:"Name"`
	PostgreSQL sqlutils.Config   `yaml:"PostgreSQL"`
	ClickHouse clickhouse.Config `yaml:"ClickHouse"`
}

var (
	instance    service.Repository
	analyticsDB service.AnalyticsRepository
	once        sync.Once
	onceCH      sync.Once
	mu          sync.RWMutex
	muCH        sync.RWMutex
)

// GetInstance returns the singleton PostgreSQL repository instance
func GetInstance(ctx context.Context, cfg *Config) service.Repository {
	mu.RLock()
	if instance != nil {
		defer mu.RUnlock()
		return instance
	}
	mu.RUnlock()

	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		if instance == nil {
			switch cfg.Name {
			case "PostgreSQL":
				if cfg.PostgreSQL.Name == "" {
					cfg.PostgreSQL.Name = databases.Postgres
				}
				repo, err := postgresql.NewRepository(ctx, &cfg.PostgreSQL)
				if err != nil {
					panic("failed to create repository: " + err.Error())
				}
				instance = repo
			default:
				panic("unsupported database type: " + cfg.Name)
			}
		}
	})

	return instance
}

// GetAnalyticsInstance returns the singleton ClickHouse analytics repository.
// Returns nil if ClickHouse is not configured.
func GetAnalyticsInstance(ctx context.Context, cfg *Config) service.AnalyticsRepository {
	if cfg.ClickHouse.DSN == "" {
		return nil
	}

	muCH.RLock()
	if analyticsDB != nil {
		defer muCH.RUnlock()
		return analyticsDB
	}
	muCH.RUnlock()

	onceCH.Do(func() {
		muCH.Lock()
		defer muCH.Unlock()
		if analyticsDB == nil {
			repo, err := clickhouse.NewRepository(ctx, &cfg.ClickHouse)
			if err != nil {
				panic("failed to create ClickHouse repository: " + err.Error())
			}
			analyticsDB = repo
		}
	})

	return analyticsDB
}
