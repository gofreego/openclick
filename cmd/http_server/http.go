package http_server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofreego/openclick/api/openclick_v1"
	"github.com/gofreego/openclick/internal/configs"
	"github.com/gofreego/openclick/internal/repository"
	"github.com/gofreego/openclick/internal/service"

	"github.com/gofreego/goutils/api"
	"github.com/gofreego/goutils/api/debug"

	"github.com/gofreego/goutils/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type HTTPServer struct {
	cfg    *configs.Configuration
	server *http.Server
}

func (a *HTTPServer) Name() string {
	return "HTTP_Server"
}

func (a *HTTPServer) Shutdown(ctx context.Context) {
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Panic(ctx, "failed to shutdown %s : %v", a.Name(), err)
	}
}

func NewHTTPServer(cfg *configs.Configuration) *HTTPServer {
	return &HTTPServer{
		cfg: cfg,
	}
}

func (a *HTTPServer) Run(ctx context.Context) error {

	if a.cfg.Server.HTTPPort == 0 {
		logger.Panic(ctx, "http port is not provided")
	}

	repo := repository.GetInstance(ctx, &a.cfg.Repository)
	analyticsDB := repository.GetAnalyticsInstance(ctx, &a.cfg.Repository)
	service := service.NewService(ctx, &a.cfg.Service, repo, analyticsDB)

	mux := runtime.NewServeMux()

	api.RegisterSwaggerHandler(ctx, mux, "/openclick/v1/swagger", "./api/docs/proto", "/openclick/v1/openclick.swagger.json")
	err := openclick_v1.RegisterBaseServiceHandlerServer(ctx, mux, service)
	if err != nil {
		logger.Panic(ctx, "failed to register ping service : %v", err)
	}

	// Register debug endpoints if enabled
	if a.cfg.Debug.Enabled {
		debug.RegisterDebugHandlersWithGateway(ctx, &a.cfg.Debug, mux, a.cfg.Logger.AppName, string(a.cfg.Logger.Build), "/openclick/v1")
	}

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.Server.HTTPPort),
		Handler: logger.WithRequestMiddleware(logger.WithRequestTimeMiddleware(api.CORSMiddleware(mux))),
	}

	logger.Info(ctx, "Starting HTTP server on port %d", a.cfg.Server.HTTPPort)
	logger.Info(ctx, "Swagger UI is available at `http://localhost:%d/openclick/v1/swagger`", a.cfg.Server.HTTPPort)
	if a.cfg.Debug.Enabled {
		logger.Info(ctx, "Debug dashboard available at `http://localhost:%d/openclick/v1/debug`", a.cfg.Server.HTTPPort)
	}
	// Start HTTP server (and proxy calls to gRPC server endpoint)
	err = a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Panic(ctx, "failed to start http server : %v", err)
	}
	return nil
}
