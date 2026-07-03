package todo

import (
	"context"
	"strings"

	todoerr "github.com/QZAiXH/kratos-layout/internal/pkg/errors/todo"

	pkgerrors "github.com/pkg/errors"
	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
)

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

// GetTodo 根据编号返回待办事项。
func (uc *UseCase) GetTodo(ctx context.Context, id string) (*Todo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, pkgerrors.WithStack(todoerr.ErrInvalidArgument)
	}
	return uc.repo.FindByID(ctx, id)
}

// ListTodos 返回待办事项列表。
func (uc *UseCase) ListTodos(ctx context.Context, opts ...ListOption) ([]*Todo, error) {
	return uc.repo.ListTodos(ctx, opts...)
}
