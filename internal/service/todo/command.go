package todo

import (
	"context"
	"strings"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"

	"go.einride.tech/aip/fieldmask"
	"google.golang.org/protobuf/types/known/emptypb"
)

// CreateTodo 创建待办事项。
func (s *Service) CreateTodo(ctx context.Context, req *v1.CreateTodoRequest) (*v1.Todo, error) {
	todo, err := s.uc.CreateTodo(ctx, convertTodo(req.GetTodo()))
	if err != nil {
		return nil, err
	}
	return convertTodoReply(todo), nil
}

// UpdateTodo 更新待办事项。
func (s *Service) UpdateTodo(ctx context.Context, req *v1.UpdateTodoRequest) (*v1.Todo, error) {
	if strings.TrimSpace(req.GetTodo().GetId()) == "" || req.GetUpdateMask() == nil || len(req.GetUpdateMask().GetPaths()) == 0 {
		return nil, todobiz.ErrTodoInvalidArgument
	}
	current, err := s.GetTodo(ctx, &v1.GetTodoRequest{Id: req.GetTodo().GetId()})
	if err != nil {
		return nil, err
	}
	fieldmask.Update(req.GetUpdateMask(), current, req.GetTodo())
	todo, err := s.uc.UpdateTodo(ctx, convertTodo(current))
	if err != nil {
		return nil, err
	}
	return convertTodoReply(todo), nil
}

// DeleteTodo 删除待办事项。
func (s *Service) DeleteTodo(ctx context.Context, req *v1.DeleteTodoRequest) (*emptypb.Empty, error) {
	if err := s.uc.DeleteTodo(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
