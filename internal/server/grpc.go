package server

import (
	"log/slog"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	"github.com/QZAiXH/kratos-layout/internal/conf"
	"github.com/QZAiXH/kratos-layout/internal/service"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	"github.com/go-kratos/kratos/v3/transport/grpc"
)

// NewGRPCServer 创建 gRPC 服务并注册业务服务。
func NewGRPCServer(c *conf.Server, security middleware.Middleware, logger *slog.Logger, todo *service.TodoService) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
			security,
		),
	}
	grpcConf := c.GetGrpc()
	if grpcConf.GetNetwork() != "" {
		opts = append(opts, grpc.Network(grpcConf.GetNetwork()))
	}
	if grpcConf.GetAddr() != "" {
		opts = append(opts, grpc.Address(grpcConf.GetAddr()))
	}
	if grpcConf.GetTimeout() != nil {
		opts = append(opts, grpc.Timeout(grpcConf.GetTimeout().AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterTodoServiceServer(srv, todo)
	return srv
}
