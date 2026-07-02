package todo

import (
	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"

	"github.com/go-kratos/kratos/v3/errors"
)

var (
	// ErrTodoNotFound 表示待办事项不存在。
	ErrTodoNotFound = errors.NotFound(v1.ErrorReason_TODO_NOT_FOUND.String(), "todo not found")
	// ErrTodoInvalidArgument 表示待办事项参数非法。
	ErrTodoInvalidArgument = errors.BadRequest(v1.ErrorReason_TODO_INVALID_ARGUMENT.String(), "invalid todo argument")
)
