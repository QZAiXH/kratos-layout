package todo

import (
	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"
)

// Service 实现待办事项 API 服务。
type Service struct {
	v1.UnimplementedTodoServiceServer // UnimplementedTodoServiceServer 保持向前兼容的 gRPC 嵌入实现。

	uc *todobiz.UseCase // uc 是待办事项业务用例。
}

// NewService 创建待办事项服务。
func NewService(uc *todobiz.UseCase) *Service {
	return &Service{uc: uc}
}
