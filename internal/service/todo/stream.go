package todo

import (
	"errors"
	"io"
	"strings"
	"time"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"

	"go.einride.tech/aip/filtering"
	"go.einride.tech/aip/ordering"
	"go.einride.tech/aip/pagination"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WatchTodos 通过服务端流发送待办事项快照。
func (s *Service) WatchTodos(req *v1.WatchTodosRequest, stream v1.TodoService_WatchTodosServer) error {
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
		return err
	}
	filter, err := filtering.ParseFilter(req, declarations)
	if err != nil {
		return err
	}
	pageToken, err := pagination.ParsePageToken(req)
	if err != nil {
		return err
	}
	orderBy, err := ordering.ParseOrderBy(req)
	if err != nil {
		return err
	}
	if err := orderBy.ValidateForPaths("id", "title", "create_time", "update_time"); err != nil {
		return err
	}
	if req.PageSize <= 0 {
		req.PageSize = defaultPageSize
	}
	todos, err := s.uc.ListTodos(stream.Context(),
		todobiz.ListFilter(filter),
		todobiz.ListOrderBy(orderBy),
		todobiz.ListLimit(int(req.PageSize)),
		todobiz.ListOffset(int(pageToken.Offset)),
	)
	if err != nil {
		return err
	}
	for _, todo := range todos {
		if err := stream.Send(newTodoEvent("snapshot", todo)); err != nil {
			return err
		}
	}
	return nil
}

// SyncTodos 通过双向流交换待办事项变更。
func (s *Service) SyncTodos(stream v1.TodoService_SyncTodosServer) error {
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var event *v1.TodoEvent
		switch strings.ToLower(req.GetAction()) {
		case "create":
			todo, err := s.CreateTodo(stream.Context(), &v1.CreateTodoRequest{Todo: req.GetTodo()})
			if err != nil {
				return err
			}
			event = newTodoEvent("created", convertTodo(todo))
		case "update":
			todo, err := s.UpdateTodo(stream.Context(), &v1.UpdateTodoRequest{
				Todo:       req.GetTodo(),
				UpdateMask: req.GetUpdateMask(),
			})
			if err != nil {
				return err
			}
			event = newTodoEvent("updated", convertTodo(todo))
		case "delete":
			id := req.GetId()
			if strings.TrimSpace(id) == "" {
				id = req.GetTodo().GetId()
			}
			if _, err := s.DeleteTodo(stream.Context(), &v1.DeleteTodoRequest{Id: id}); err != nil {
				return err
			}
			event = &v1.TodoEvent{
				Action:    "deleted",
				Todo:      &v1.Todo{Id: id},
				EventTime: timestamppb.Now(),
				Type:      v1.TodoEventType_TODO_EVENT_TYPE_DELETED,
			}
		default:
			return todobiz.ErrTodoInvalidArgument
		}
		if err := stream.Send(event); err != nil {
			return err
		}
	}
}

// newTodoEvent 创建待办事项流式事件。
func newTodoEvent(action string, todo *todobiz.Todo) *v1.TodoEvent {
	return &v1.TodoEvent{
		Action:    action,
		Todo:      convertTodoReply(todo),
		EventTime: timestamppb.New(time.Now()),
		Type:      todoEventType(action),
	}
}

// todoEventType 将事件动作转换为枚举类型。
func todoEventType(action string) v1.TodoEventType {
	switch strings.ToLower(action) {
	case "created", "create":
		return v1.TodoEventType_TODO_EVENT_TYPE_CREATED
	case "updated", "update":
		return v1.TodoEventType_TODO_EVENT_TYPE_UPDATED
	case "deleted", "delete":
		return v1.TodoEventType_TODO_EVENT_TYPE_DELETED
	case "snapshot":
		return v1.TodoEventType_TODO_EVENT_TYPE_SNAPSHOT
	default:
		return v1.TodoEventType_TODO_EVENT_TYPE_UNSPECIFIED
	}
}
