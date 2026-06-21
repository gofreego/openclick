package repository

import (
	"context"
	"sync"

	sqlutils "github.com/gofreego/goutils/databases/connections/sql"
	"github.com/gofreego/openclick/internal/repository/memory"
	"github.com/gofreego/openclick/internal/repository/postgresql"
	"github.com/gofreego/openclick/internal/service"
)

type Config struct {
	Name       string          `yaml:"Name"`
	Memory     memory.Config   `yaml:"Memory"`
	PostgreSQL sqlutils.Config `yaml:"PostgreSQL"`
}

var (
	instance service.Repository
	once     sync.Once
	mu       sync.RWMutex
)

// GetInstance returns the singleton instance of the repository
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
			case "Memory":
				repo, err := memory.NewRepository(ctx, &cfg.Memory)
				if err != nil {
					panic("failed to create repository: " + err.Error())
				}
				instance = repo
			case "PostgreSQL":
				repo, err := postgresql.NewRepository(ctx, &cfg.PostgreSQL)
				if err != nil {
					panic("failed to create repository: " + err.Error())
				}
				instance = repo
			default:
				panic("unknown repository name: " + cfg.Name)
			}
		}
	})

	return instance
}
