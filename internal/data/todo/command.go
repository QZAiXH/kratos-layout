package todo

import (
	"context"
	"time"

	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"
	todoerr "github.com/QZAiXH/kratos-layout/internal/pkg/errors/todo"
	"github.com/QZAiXH/kratos-layout/internal/pkg/id"

	pkgerrors "github.com/pkg/errors"
)

// CreateTodo 创建待办事项并分配内存编号。
func (r *todoRepo) CreateTodo(_ context.Context, todo *todobiz.Todo) (*todobiz.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	todo = cloneTodo(todo)
	todo.ID = id.ULID("")
	todo.CreateTime = now
	todo.UpdateTime = now
	r.todos[todo.ID] = cloneTodo(todo)
	return cloneTodo(todo), nil
}

// UpdateTodo 更新已存在的待办事项。
func (r *todoRepo) UpdateTodo(_ context.Context, todo *todobiz.Todo) (*todobiz.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, ok := r.todos[todo.ID]
	if !ok {
		return nil, pkgerrors.WithStack(todoerr.ErrNotFound)
	}
	updated := cloneTodo(todo)
	updated.CreateTime = current.CreateTime
	updated.UpdateTime = time.Now()
	r.todos[updated.ID] = cloneTodo(updated)
	return cloneTodo(updated), nil
}

// DeleteTodo 删除待办事项。
func (r *todoRepo) DeleteTodo(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.todos[id]; !ok {
		return pkgerrors.WithStack(todoerr.ErrNotFound)
	}
	delete(r.todos, id)
	return nil
}
