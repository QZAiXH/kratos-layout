package todo

import (
	"testing"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	pkgerrors "github.com/pkg/errors"
)

// TestTodoErrors 验证待办事项错误契约保持稳定。
func TestTodoErrors(t *testing.T) {
	notFound := kratoserrors.FromError(ErrNotFound)
	if notFound.Code != 404 || notFound.Reason != "TODO_NOT_FOUND" {
		t.Fatalf("ErrNotFound = %+v, want 404 TODO_NOT_FOUND", notFound)
	}
	if !IsNotFound(ErrNotFound) {
		t.Fatal("IsNotFound() = false, want true")
	}
	if !IsNotFound(pkgerrors.WithStack(ErrNotFound)) {
		t.Fatal("IsNotFound(wrapped) = false, want true")
	}

	invalid := kratoserrors.FromError(ErrInvalidArgument)
	if invalid.Code != 400 || invalid.Reason != "TODO_INVALID_ARGUMENT" {
		t.Fatalf("ErrInvalidArgument = %+v, want 400 TODO_INVALID_ARGUMENT", invalid)
	}
	if !IsInvalidArgument(ErrInvalidArgument) {
		t.Fatal("IsInvalidArgument() = false, want true")
	}
	if !IsInvalidArgument(pkgerrors.WithStack(ErrInvalidArgument)) {
		t.Fatal("IsInvalidArgument(wrapped) = false, want true")
	}
}
