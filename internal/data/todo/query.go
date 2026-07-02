package todo

import (
	"cmp"
	"context"
	"slices"

	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"
)

// FindByID 根据编号查询待办事项。
func (r *todoRepo) FindByID(_ context.Context, id string) (*todobiz.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todo, ok := r.todos[id]
	if !ok {
		return nil, todobiz.ErrTodoNotFound
	}
	return cloneTodo(todo), nil
}

// ListTodos 按查询选项返回待办事项列表。
func (r *todoRepo) ListTodos(_ context.Context, opts ...todobiz.ListOption) ([]*todobiz.Todo, error) {
	options := todobiz.ListOptions{Limit: 20}
	for _, opt := range opts {
		opt(&options)
	}
	if options.Offset < 0 || options.Limit <= 0 {
		return nil, todobiz.ErrTodoInvalidArgument
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	todos := make([]*todobiz.Todo, 0, len(r.todos))
	for _, todo := range r.todos {
		todos = append(todos, cloneTodo(todo))
	}
	slices.SortFunc(todos, func(a, b *todobiz.Todo) int {
		return cmp.Or(
			cmp.Compare(a.CreateTime.UnixNano(), b.CreateTime.UnixNano()),
			cmp.Compare(a.ID, b.ID),
		)
	})

	if options.Offset >= len(todos) {
		return []*todobiz.Todo{}, nil
	}
	end := options.Offset + options.Limit
	if end > len(todos) {
		end = len(todos)
	}
	return todos[options.Offset:end], nil
}
