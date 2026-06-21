package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"net/http"

	"github.com/gofreego/openclick/cmd/grpc_server"
	"github.com/gofreego/openclick/cmd/http_server"
	"github.com/gofreego/openclick/internal/configs"
	"github.com/gofreego/openclick/internal/constants"

	"github.com/gofreego/goutils/apputils"
	"github.com/gofreego/goutils/logger"
)

var (
	env  string
	path string
)

//go:embed all:dashboard-ui/dist
var uiDist embed.FS

func getUIFileSystem() http.FileSystem {
	// Re-map the embedded filesystem to the root of 'dist'
	fsys, err := fs.Sub(uiDist, "dashboard-ui/dist")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

func getIndexHTML() []byte {
	data, err := fs.ReadFile(uiDist, "dashboard-ui/dist/index.html")
	if err != nil {
		panic(err)
	}
	return data
}

func getUIHandler() http.Handler {
	return http_server.GetUIHandler(getUIFileSystem(), getIndexHTML())
}

func main() {
	flag.StringVar(&env, "env", "dev", "-env=dev")
	flag.StringVar(&path, "path", ".", "-path=./")
	flag.Parse()
	ctx := context.Background()

	conf := configs.LoadConfig(ctx, path, env)

	conf.Logger.InitiateLogger()
	logger.AddMiddleLayers(logger.RequestMiddleLayer)

	// starting application
	var apps []apputils.Application
	for _, appName := range conf.AppNames {
		switch appName {
		case constants.HTTP_SERVER:
			apps = append(apps, http_server.NewHTTPServer(conf, getUIHandler()))
		case constants.GRPC_SERVER:
			apps = append(apps, grpc_server.NewGRPCServer(conf))
		default:
			logger.Panic(ctx, "invalid application name provided `%s`", appName)
		}
	}

	for _, app := range apps {
		logger.Info(ctx, "Starting %s", app.Name())
		go app.Run(ctx)
	}

	apputils.GracefulShutdown(ctx, apps...)
}
