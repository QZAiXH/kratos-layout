package server

import (
	"context"
	"testing"

	v1 "helloworld/api/todo/v1"
)

// TestNewProtectedMatcher 验证安全匹配器只放行模板白名单操作。
func TestNewProtectedMatcher(t *testing.T) {
	match := NewProtectedMatcher()
	if match(context.Background(), v1.OperationTodoServiceCreateTodo) {
		t.Fatal("CreateTodo protected = true, want false")
	}
	if !match(context.Background(), "/private.Service/Call") {
		t.Fatal("private operation protected = false, want true")
	}
}
