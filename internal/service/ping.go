package service

import (
	"context"

	"github.com/gofreego/goutils/logger"
	"github.com/gofreego/openclick/api/openclick_v1"
)

func (s *Service) Ping(ctx context.Context, req *openclick_v1.PingRequest) (*openclick_v1.PingResponse, error) {
	logger.Debug(ctx, "Ping request received, %v", req.Message)
	err := s.repo.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return &openclick_v1.PingResponse{
		Message: "Its fine here...!",
	}, nil
}
