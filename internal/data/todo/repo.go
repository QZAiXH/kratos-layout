package todo

import (
	"cmp"
	"context"
	"slices"
	"sync"
	"time"

	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"
	"github.com/QZAiXH/kratos-layout/internal/data/base"
	"github.com/QZAiXH/kratos-layout/internal/pkg/id"
)

// todoRepo 用内存存储实现待办事项仓储。
type todoRepo struct {
	data *base.Data // data 是数据层共享依赖。

	mu    sync.RWMutex             // mu 保护内存待办事项集合。
	todos map[string]*todobiz.Todo // todos 保存待办事项快照。
}

// NewRepo 创建待办事项仓储实例。
func NewRepo(data *base.Data) todobiz.Repo {
	return &todoRepo{
		data:  data,
		todos: make(map[string]*todobiz.Todo),
	}
}

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
		return nil, todobiz.ErrTodoNotFound
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
		return todobiz.ErrTodoNotFound
	}
	delete(r.todos, id)
	return nil
}

// cloneTodo 复制待办事项以避免外部修改内存状态。
func cloneTodo(todo *todobiz.Todo) *todobiz.Todo {
	if todo == nil {
		return nil
	}
	return &todobiz.Todo{
		ID:         todo.ID,
		Title:      todo.Title,
		Content:    todo.Content,
		Completed:  todo.Completed,
		CreateTime: todo.CreateTime,
		UpdateTime: todo.UpdateTime,
	}
}
