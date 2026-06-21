package service

import (
	"context"

	"github.com/gofreego/openclick/api/openclick_v1"
)

type Config struct {
}

type Repository interface {
	Ping(ctx context.Context) error
}

type Service struct {
	repo Repository
	openclick_v1.UnimplementedBaseServiceServer
}

func NewService(ctx context.Context, cfg *Config, repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}
