package main

import (
	"flag"
	"log/slog"
	"os"

	"helloworld/internal/conf"
	"helloworld/internal/pkg/zaplog"

	"github.com/go-kratos/kratos/contrib/otel/v3/tracing"
	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/file"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	logger, cleanupLogger, err := newLogger(bc.Log)
	if err != nil {
		panic(err)
	}
	defer cleanupLogger()
	log.SetDefault(logger)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Auth, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func newLogger(c *conf.Log) (*slog.Logger, func() error, error) {
	var opts []zaplog.Option
	if c != nil {
		opts = append(opts, zaplog.WithLevel(c.Level))
		if c.Filepath != "" {
			opts = append(opts, zaplog.WithFilePath(c.Filepath))
			if c.Rotate != nil {
				opts = append(opts, zaplog.WithRotate(
					int(c.Rotate.MaxSize),
					int(c.Rotate.MaxBackups),
					int(c.Rotate.MaxAge),
					c.Rotate.Compress,
				))
			}
		}
	}

	handler, cleanup, err := zaplog.NewHandler(opts...)
	if err != nil {
		return nil, nil, err
	}
	logger := log.NewLogger(handler, log.WithExtractor(tracing.TraceAttrs)).With(
		slog.String("service.id", id),
		slog.String("service.name", Name),
		slog.String("service.version", Version),
	)
	return logger, cleanup, nil
}
