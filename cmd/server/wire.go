//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/QZAiXH/kratos-layout/internal/biz"
	"github.com/QZAiXH/kratos-layout/internal/conf"
	"github.com/QZAiXH/kratos-layout/internal/data"
	"github.com/QZAiXH/kratos-layout/internal/dep"
	"github.com/QZAiXH/kratos-layout/internal/job"
	"github.com/QZAiXH/kratos-layout/internal/server"
	"github.com/QZAiXH/kratos-layout/internal/service"

	"github.com/go-kratos/kratos/v3"
	"github.com/google/wire"
)

// wireApp 通过 Wire 初始化 Kratos 应用依赖图。
func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Job, *slog.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(dep.ProviderSet, job.ProviderSet, server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
