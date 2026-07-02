package todo

import (
	"context"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"

	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
	"go.einride.tech/aip/pagination"
)

const defaultPageSize = 20

// GetTodo 根据编号返回待办事项。
func (s *Service) GetTodo(ctx context.Context, req *v1.GetTodoRequest) (*v1.Todo, error) {
	todo, err := s.uc.GetTodo(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return convertTodoReply(todo), nil
}

// ListTodos 返回待办事项列表。
func (s *Service) ListTodos(ctx context.Context, req *v1.ListTodosRequest) (*v1.TodoSet, error) {
	declarations, err := filtering.NewDeclarations(
		filtering.DeclareStandardFunctions(),
		filtering.DeclareIdent("id", filtering.TypeString),
		filtering.DeclareIdent("title", filtering.TypeString),
		filtering.DeclareIdent("content", filtering.TypeString),
		filtering.DeclareIdent("completed", filtering.TypeBool),
		filtering.DeclareIdent("create_time", filtering.TypeTimestamp),
		filtering.DeclareIdent("update_time", filtering.TypeTimestamp),
	)
	if err != nil {
		return nil, err
	}
	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		return nil, err
	}
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return nil, err
	}
	orderBy, err := ordering.ParseOrderBy(req)
	if err != nil {
		return nil, err
	}
	if err := orderBy.ValidateForPaths("id", "title", "create_time", "update_time"); err != nil {
		return nil, err
	}
	if req.PageSize <= 0 {
		req.PageSize = defaultPageSize
	}
	todos, err := s.uc.ListTodos(ctx,
		todobiz.ListFilter(filter),
		todobiz.ListOrderBy(orderBy),
		todobiz.ListLimit(int(req.PageSize)),
		todobiz.ListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return nil, err
	}
	set := &v1.TodoSet{
		Todos: make([]*v1.Todo, 0, len(todos)),
	}
	if len(todos) >= int(req.PageSize) {
		set.NextPageToken = pageToken.Next(req).String()
	}
	for _, todo := range todos {
		set.Todos = append(set.Todos, convertTodoReply(todo))
	}
	return set, nil
}
