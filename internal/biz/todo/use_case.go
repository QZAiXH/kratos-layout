package todo

import (
	"context"
)

// Repo 定义待办事项仓储接口。
type Repo interface {
	// FindByID 根据编号查询待办事项。
	FindByID(context.Context, string) (*Todo, error)
	// ListTodos 按查询选项列出待办事项。
	ListTodos(context.Context, ...ListOption) ([]*Todo, error)
	// CreateTodo 创建待办事项。
	CreateTodo(context.Context, *Todo) (*Todo, error)
	// UpdateTodo 更新待办事项。
	UpdateTodo(context.Context, *Todo) (*Todo, error)
	// DeleteTodo 删除待办事项。
	DeleteTodo(context.Context, string) error
}

// UseCase 编排待办事项业务流程。
type UseCase struct {
	repo Repo // repo 是待办事项仓储接口。
}

// NewUseCase 创建待办事项用例。
func NewUseCase(repo Repo) *UseCase {
	return &UseCase{repo: repo}
}
