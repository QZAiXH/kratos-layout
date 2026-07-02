package todo

import (
	"context"
	"io"
	"testing"

	v1 "github.com/QZAiXH/kratos-layout/api/todo/v1"
	todobiz "github.com/QZAiXH/kratos-layout/internal/biz/todo"
	"github.com/QZAiXH/kratos-layout/internal/data/base"
	tododata "github.com/QZAiXH/kratos-layout/internal/data/todo"

	kratoserrors "github.com/go-kratos/kratos/v3/errors"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

// newTestTodoService 创建使用内存仓储的测试服务。
func newTestTodoService() *Service {
	repo := tododata.NewRepo(&base.Data{})
	uc := todobiz.NewUseCase(repo)
	return NewService(uc)
}

// TestTodoServiceCRUD 验证 Todo 服务完整创建、读取、更新和删除流程。
func TestTodoServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()

	created, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{
		Todo: &v1.Todo{
			Title:     "write tests",
			Content:   "cover todo CRUD",
			Completed: false,
		},
	})
	if err != nil {
		t.Fatalf("CreateTodo() error = %v", err)
	}
	if created.GetId() == "" {
		t.Fatal("CreateTodo() id is empty")
	}
	if created.GetCreateTime() == nil || created.GetUpdateTime() == nil {
		t.Fatal("CreateTodo() did not set timestamps")
	}

	got, err := svc.GetTodo(ctx, &v1.GetTodoRequest{Id: created.GetId()})
	if err != nil {
		t.Fatalf("GetTodo() error = %v", err)
	}
	if got.GetTitle() != "write tests" || got.GetContent() != "cover todo CRUD" {
		t.Fatalf("GetTodo() = %+v, want created todo", got)
	}

	updated, err := svc.UpdateTodo(ctx, &v1.UpdateTodoRequest{
		Todo: &v1.Todo{
			Id:        created.GetId(),
			Title:     "write service tests",
			Completed: true,
		},
		UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"title", "completed"}},
	})
	if err != nil {
		t.Fatalf("UpdateTodo() error = %v", err)
	}
	if updated.GetTitle() != "write service tests" || !updated.GetCompleted() {
		t.Fatalf("UpdateTodo() = %+v, want updated title and completed", updated)
	}
	if updated.GetContent() != "cover todo CRUD" {
		t.Fatalf("UpdateTodo() content = %q, want original content", updated.GetContent())
	}

	if _, err := svc.DeleteTodo(ctx, &v1.DeleteTodoRequest{Id: created.GetId()}); err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if _, err := svc.GetTodo(ctx, &v1.GetTodoRequest{Id: created.GetId()}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("GetTodo() after delete error = %v, want not found", err)
	}
}

// TestTodoServiceListTodosPagination 验证列表接口按游标分页返回下一页令牌。
func TestTodoServiceListTodosPagination(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()

	for _, title := range []string{"first", "second", "third"} {
		if _, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{Todo: &v1.Todo{Title: title}}); err != nil {
			t.Fatalf("CreateTodo(%q) error = %v", title, err)
		}
	}

	firstPage, err := svc.ListTodos(ctx, &v1.ListTodosRequest{PageSize: 2})
	if err != nil {
		t.Fatalf("ListTodos(first page) error = %v", err)
	}
	if len(firstPage.GetTodos()) != 2 {
		t.Fatalf("ListTodos(first page) len = %d, want 2", len(firstPage.GetTodos()))
	}
	if firstPage.GetNextPageToken() == "" {
		t.Fatal("ListTodos(first page) next_page_token is empty")
	}

	secondPage, err := svc.ListTodos(ctx, &v1.ListTodosRequest{
		PageSize:  2,
		PageToken: firstPage.GetNextPageToken(),
	})
	if err != nil {
		t.Fatalf("ListTodos(second page) error = %v", err)
	}
	if len(secondPage.GetTodos()) != 1 {
		t.Fatalf("ListTodos(second page) len = %d, want 1", len(secondPage.GetTodos()))
	}
	if secondPage.GetNextPageToken() != "" {
		t.Fatalf("ListTodos(second page) next_page_token = %q, want empty", secondPage.GetNextPageToken())
	}
	if secondPage.GetTodos()[0].GetTitle() != "third" {
		t.Fatalf("ListTodos(second page) title = %q, want third", secondPage.GetTodos()[0].GetTitle())
	}
}

// TestTodoServiceListTodosFilterAndOrderByValidation 验证列表筛选和排序参数会进入业务查询路径。
func TestTodoServiceListTodosFilterAndOrderByValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()

	for _, todo := range []*v1.Todo{
		{Title: "write docs", Content: "public docs", Completed: true},
		{Title: "fix bug", Content: "private bug", Completed: false},
		{Title: "fix api", Content: "public api", Completed: true},
	} {
		if _, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{Todo: todo}); err != nil {
			t.Fatalf("CreateTodo(%q) error = %v", todo.GetTitle(), err)
		}
	}

	reply, err := svc.ListTodos(ctx, &v1.ListTodosRequest{
		PageSize: 10,
		Filter:   `title:"fix" AND completed`,
		OrderBy:  "title desc",
	})
	if err != nil {
		t.Fatalf("ListTodos(filter/order) error = %v", err)
	}
	if len(reply.GetTodos()) != 3 {
		t.Fatalf("ListTodos(filter/order) len = %d, want 3", len(reply.GetTodos()))
	}
	if reply.GetTodos()[0].GetTitle() != "write docs" {
		t.Fatalf("ListTodos(filter/order) first title = %q, want ID order", reply.GetTodos()[0].GetTitle())
	}
}

// TestTodoServiceValidation 验证服务层会透传业务层参数校验和未找到错误。
func TestTodoServiceValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()

	if _, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{Todo: &v1.Todo{Title: " "}}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("CreateTodo(empty title) error = %v, want bad request", err)
	}
	if _, err := svc.UpdateTodo(ctx, &v1.UpdateTodoRequest{
		Todo:       &v1.Todo{Id: "01JZ4T00000000000000000000", Title: "missing mask"},
		UpdateMask: &fieldmaskpb.FieldMask{},
	}); !kratoserrors.IsBadRequest(err) {
		t.Fatalf("UpdateTodo(empty mask) error = %v, want bad request", err)
	}
	if _, err := svc.ListTodos(ctx, &v1.ListTodosRequest{PageToken: "bad-token"}); err == nil {
		t.Fatal("ListTodos(bad token) error = nil, want error")
	}
	if _, err := svc.ListTodos(ctx, &v1.ListTodosRequest{Filter: `unknown:"value"`}); err == nil {
		t.Fatal("ListTodos(unsupported filter) error = nil, want error")
	}
	if _, err := svc.ListTodos(ctx, &v1.ListTodosRequest{OrderBy: "content"}); err == nil {
		t.Fatal("ListTodos(unsupported order_by) error = nil, want error")
	}
	if _, err := svc.DeleteTodo(ctx, &v1.DeleteTodoRequest{Id: "01JZ4T00000000000000000000"}); !kratoserrors.IsNotFound(err) {
		t.Fatalf("DeleteTodo(missing id) error = %v, want not found", err)
	}
}

// TestTodoServiceWatchTodos 验证服务端流接口会发送当前 Todo 快照事件。
func TestTodoServiceWatchTodos(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()

	for _, todo := range []*v1.Todo{
		{Title: "open task", Completed: false},
		{Title: "done task", Completed: true},
	} {
		if _, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{Todo: todo}); err != nil {
			t.Fatalf("CreateTodo(%q) error = %v", todo.GetTitle(), err)
		}
	}

	stream := &watchTodosStream{fakeServerStream: fakeServerStream{ctx: ctx}}
	if err := svc.WatchTodos(&v1.WatchTodosRequest{
		PageSize: 10,
	}, stream); err != nil {
		t.Fatalf("WatchTodos() error = %v", err)
	}
	if len(stream.events) != 2 {
		t.Fatalf("WatchTodos() events len = %d, want 2", len(stream.events))
	}
	if stream.events[0].GetAction() != "snapshot" || stream.events[0].GetTodo().GetTitle() != "open task" {
		t.Fatalf("WatchTodos() event = %+v, want open snapshot", stream.events[0])
	}
}

// TestTodoServiceSyncTodos 验证双向流接口按 create/update/delete 请求顺序返回事件。
func TestTodoServiceSyncTodos(t *testing.T) {
	ctx := context.Background()
	svc := newTestTodoService()
	existing, err := svc.CreateTodo(ctx, &v1.CreateTodoRequest{
		Todo: &v1.Todo{Title: "existing todo", Content: "before stream"},
	})
	if err != nil {
		t.Fatalf("CreateTodo(existing) error = %v", err)
	}
	stream := &syncTodosStream{
		fakeServerStream: fakeServerStream{ctx: ctx},
		requests: []*v1.SyncTodoRequest{
			{
				Action: "create",
				Todo:   &v1.Todo{Title: "streamed todo", Content: "from bidi stream"},
			},
			{
				Action: "update",
				Todo:   &v1.Todo{Id: existing.GetId(), Completed: true},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"completed"},
				},
			},
			{
				Action: "delete",
				Id:     existing.GetId(),
			},
		},
	}

	if err := svc.SyncTodos(stream); err != nil {
		t.Fatalf("SyncTodos() error = %v", err)
	}
	if got := len(stream.events); got != 3 {
		t.Fatalf("SyncTodos() events len = %d, want 3", got)
	}
	if stream.events[0].GetAction() != "created" || stream.events[0].GetTodo().GetId() == "" {
		t.Fatalf("SyncTodos() create event = %+v, want created id", stream.events[0])
	}
	if stream.events[1].GetAction() != "updated" || !stream.events[1].GetTodo().GetCompleted() {
		t.Fatalf("SyncTodos() update event = %+v, want completed update", stream.events[1])
	}
	if stream.events[2].GetAction() != "deleted" || stream.events[2].GetTodo().GetId() != existing.GetId() {
		t.Fatalf("SyncTodos() delete event = %+v, want deleted id %s", stream.events[2], existing.GetId())
	}
}

// fakeServerStream 提供测试用 gRPC ServerStream 基础实现。
type fakeServerStream struct {
	ctx context.Context // ctx 是测试流上下文。
}

// SetHeader 记录测试流响应头。
func (s fakeServerStream) SetHeader(metadata.MD) error { return nil }

// SendHeader 发送测试流响应头。
func (s fakeServerStream) SendHeader(metadata.MD) error { return nil }

// SetTrailer 设置测试流尾部元数据。
func (s fakeServerStream) SetTrailer(metadata.MD) {}

// Context 返回测试流上下文。
func (s fakeServerStream) Context() context.Context { return s.ctx }

// SendMsg 发送测试流原始消息。
func (s fakeServerStream) SendMsg(any) error { return nil }

// RecvMsg 接收测试流原始消息。
func (s fakeServerStream) RecvMsg(any) error { return nil }

// watchTodosStream 捕获服务端流发送的待办事项事件。
type watchTodosStream struct {
	fakeServerStream                 // fakeServerStream 提供基础流方法。
	events           []*v1.TodoEvent // events 保存已发送事件。
}

// Send 记录服务端流事件。
func (s *watchTodosStream) Send(event *v1.TodoEvent) error {
	s.events = append(s.events, event)
	return nil
}

// syncTodosStream 提供双向流测试输入并捕获输出事件。
type syncTodosStream struct {
	fakeServerStream                       // fakeServerStream 提供基础流方法。
	requests         []*v1.SyncTodoRequest // requests 是待接收的客户端请求。
	events           []*v1.TodoEvent       // events 是服务端发送的事件。
}

// Recv 返回下一条双向流请求。
func (s *syncTodosStream) Recv() (*v1.SyncTodoRequest, error) {
	if len(s.requests) == 0 {
		return nil, io.EOF
	}
	req := s.requests[0]
	s.requests = s.requests[1:]
	return req, nil
}

// Send 记录双向流响应事件。
func (s *syncTodosStream) Send(event *v1.TodoEvent) error {
	s.events = append(s.events, event)
	return nil
}
