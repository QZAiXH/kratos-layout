package todo

import (
	"context"
	"strings"
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
		return nil, ErrTodoInvalidArgument
	}
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.UpdateTodo(ctx, todo)
}

// DeleteTodo 删除待办事项。
func (uc *UseCase) DeleteTodo(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrTodoInvalidArgument
	}
	return uc.repo.DeleteTodo(ctx, id)
}
