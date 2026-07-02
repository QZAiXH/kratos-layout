package todo

import (
	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// convertTodo 将 API 待办事项转换为业务对象。
func convertTodo(in *v1.Todo) *todobiz.Todo {
	if in == nil {
		return nil
	}
	return &todobiz.Todo{
		ID:        in.GetId(),
		Title:     in.GetTitle(),
		Content:   in.GetContent(),
		Completed: in.GetCompleted(),
	}
}

// convertTodoReply 将业务对象转换为 API 待办事项。
func convertTodoReply(in *todobiz.Todo) *v1.Todo {
	if in == nil {
		return nil
	}
	return &v1.Todo{
		Id:         in.ID,
		Title:      in.Title,
		Content:    in.Content,
		Completed:  in.Completed,
		CreateTime: timestamppb.New(in.CreateTime),
		UpdateTime: timestamppb.New(in.UpdateTime),
	}
}
