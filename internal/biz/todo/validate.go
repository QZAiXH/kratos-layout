package todo

import (
	"strings"

	todoerr "github.com/QZAiXH/kratos-layout/internal/pkg/errors/todo"

	pkgerrors "github.com/pkg/errors"
)

// validateTodo 校验待办事项业务对象。
func validateTodo(todo *Todo) error {
	if todo == nil || strings.TrimSpace(todo.Title) == "" {
		return pkgerrors.WithStack(todoerr.ErrInvalidArgument)
	}
	return nil
}
