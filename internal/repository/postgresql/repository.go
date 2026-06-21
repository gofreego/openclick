package postgresql

import (
	"context"

	sqlutils "github.com/gofreego/goutils/databases/connections/sql"
)

// Repository implements the service.Repository interface using PostgreSQL
type Repository struct {
	connManager sqlutils.DBManager
}

// NewRepository creates a new PostgreSQL repository instance
func NewRepository(ctx context.Context, cfg *sqlutils.Config) (*Repository, error) {
	connManager, err := sqlutils.NewDBManager(cfg)
	if err != nil {
		return nil, err
	}
	return &Repository{connManager: connManager}, nil
}

// Ping checks the database connection
func (r *Repository) Ping(ctx context.Context) error {
	return r.connManager.Primary().Ping()
}
