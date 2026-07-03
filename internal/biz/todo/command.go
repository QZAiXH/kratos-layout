package todo

import (
	"context"
	"strings"

	todoerr "github.com/QZAiXH/kratos-layout/internal/pkg/errors/todo"

	pkgerrors "github.com/pkg/errors"
)

// CreateTodo 创建待办事项。
func (uc *UseCase) CreateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.CreateTodo(ctx, todo)
}

// UpdateTodo 更新待办事项。
func (uc *UseCase) UpdateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if todo == nil || strings.TrimSpace(todo.ID) == "" {
		return nil, pkgerrors.WithStack(todoerr.ErrInvalidArgument)
	}
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.UpdateTodo(ctx, todo)
}

// DeleteTodo 删除待办事项。
func (uc *UseCase) DeleteTodo(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return pkgerrors.WithStack(todoerr.ErrInvalidArgument)
	}
	return uc.repo.DeleteTodo(ctx, id)
}
