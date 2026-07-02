package biz

import (
	"context"
	"strings"
	"time"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"

	"github.com/go-kratos/kratos/v3/errors"
	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
)

var (
	// ErrTodoNotFound is returned when a todo does not exist.
	ErrTodoNotFound = errors.NotFound(v1.ErrorReason_TODO_NOT_FOUND.String(), "todo not found")
	// ErrTodoInvalidArgument is returned when a todo request is invalid.
	ErrTodoInvalidArgument = errors.BadRequest(v1.ErrorReason_TODO_INVALID_ARGUMENT.String(), "invalid todo argument")
)

// Todo 表示待办事项业务对象。
type Todo struct {
	ID         int64     // ID 是待办事项唯一编号。
	Title      string    // Title 是待办事项标题。
	Content    string    // Content 是待办事项内容。
	Completed  bool      // Completed 表示待办事项是否完成。
	CreateTime time.Time // CreateTime 是创建时间。
	UpdateTime time.Time // UpdateTime 是更新时间。
}

// TodoRepo 定义待办事项仓储接口。
type TodoRepo interface {
	// FindByID 根据编号查询待办事项。
	FindByID(context.Context, int64) (*Todo, error)
	// ListTodos 按查询选项列出待办事项。
	ListTodos(context.Context, ...ListOption) ([]*Todo, error)
	// CreateTodo 创建待办事项。
	CreateTodo(context.Context, *Todo) (*Todo, error)
	// UpdateTodo 更新待办事项。
	UpdateTodo(context.Context, *Todo) (*Todo, error)
	// DeleteTodo 删除待办事项。
	DeleteTodo(context.Context, int64) error
}

// ListOption configures todo list queries.
type ListOption func(*ListOptions)

// ListOptions 表示待办事项列表查询选项。
type ListOptions struct {
	Filter  filtering.Filter // Filter 是 AIP 标准过滤条件。
	OrderBy ordering.OrderBy // OrderBy 是 AIP 标准排序条件。
	Offset  int              // Offset 是分页偏移量。
	Limit   int              // Limit 是分页数量上限。
}

// ListFilter 设置 AIP 标准过滤条件。
func ListFilter(filter filtering.Filter) ListOption {
	return func(o *ListOptions) {
		o.Filter = filter
	}
}

// ListOrderBy 设置 AIP 标准排序条件。
func ListOrderBy(orderBy ordering.OrderBy) ListOption {
	return func(o *ListOptions) {
		o.OrderBy = orderBy
	}
}

// ListOffset 设置分页偏移量。
func ListOffset(offset int) ListOption {
	return func(o *ListOptions) {
		o.Offset = offset
	}
}

// ListLimit 设置分页数量上限。
func ListLimit(limit int) ListOption {
	return func(o *ListOptions) {
		o.Limit = limit
	}
}

// TodoUsecase 编排待办事项业务流程。
type TodoUsecase struct {
	repo TodoRepo // repo 是待办事项仓储接口。
}

// NewTodoUsecase 创建待办事项用例。
func NewTodoUsecase(repo TodoRepo) *TodoUsecase {
	return &TodoUsecase{repo: repo}
}

// GetTodo 根据编号返回待办事项。
func (uc *TodoUsecase) GetTodo(ctx context.Context, id int64) (*Todo, error) {
	return uc.repo.FindByID(ctx, id)
}

// ListTodos 返回待办事项列表。
func (uc *TodoUsecase) ListTodos(ctx context.Context, opts ...ListOption) ([]*Todo, error) {
	return uc.repo.ListTodos(ctx, opts...)
}

// CreateTodo 创建待办事项。
func (uc *TodoUsecase) CreateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.CreateTodo(ctx, todo)
}

// UpdateTodo 更新待办事项。
func (uc *TodoUsecase) UpdateTodo(ctx context.Context, todo *Todo) (*Todo, error) {
	if todo == nil || todo.ID <= 0 {
		return nil, ErrTodoInvalidArgument
	}
	if err := validateTodo(todo); err != nil {
		return nil, err
	}
	return uc.repo.UpdateTodo(ctx, todo)
}

// DeleteTodo 删除待办事项。
func (uc *TodoUsecase) DeleteTodo(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrTodoInvalidArgument
	}
	return uc.repo.DeleteTodo(ctx, id)
}

// validateTodo 校验待办事项业务对象。
func validateTodo(todo *Todo) error {
	if todo == nil || strings.TrimSpace(todo.Title) == "" {
		return ErrTodoInvalidArgument
	}
	return nil
}
