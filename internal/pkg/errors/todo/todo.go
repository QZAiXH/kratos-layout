package todo

import v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"

var (
	// ErrNotFound 表示待办事项不存在。
	ErrNotFound = v1.ErrorTodoNotFound("todo not found")
	// ErrInvalidArgument 表示待办事项参数非法。
	ErrInvalidArgument = v1.ErrorTodoInvalidArgument("invalid todo argument")
)

// IsNotFound 判断错误是否为待办事项不存在。
func IsNotFound(err error) bool {
	return v1.IsTodoNotFound(err)
}

// IsInvalidArgument 判断错误是否为待办事项参数非法。
func IsInvalidArgument(err error) bool {
	return v1.IsTodoInvalidArgument(err)
}
