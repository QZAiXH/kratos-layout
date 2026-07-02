package todo

import todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"

// cloneTodo 复制待办事项以避免外部修改内存状态。
func cloneTodo(todo *todobiz.Todo) *todobiz.Todo {
	if todo == nil {
		return nil
	}
	return &todobiz.Todo{
		ID:         todo.ID,
		Title:      todo.Title,
		Content:    todo.Content,
		Completed:  todo.Completed,
		CreateTime: todo.CreateTime,
		UpdateTime: todo.UpdateTime,
	}
}
