package typecatch

import (
	"testing"
	"time"
)

// sourceUser 表示测试用源结构体。
type sourceUser struct {
	Name     *string    // Name 是可空名称。
	JoinedAt *time.Time // JoinedAt 是可空加入时间。
}

// targetUser 表示测试用目标结构体。
type targetUser struct {
	Name     string    // Name 是目标名称。
	JoinedAt time.Time // JoinedAt 是目标加入时间。
}

// TestCopyTo 验证同名字段复制会应用默认指针到值类型转换。
func TestCopyTo(t *testing.T) {
	name := "alice"
	joinedAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)

	got, err := CopyTo[sourceUser, targetUser](&sourceUser{
		Name:     &name,
		JoinedAt: &joinedAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != name {
		t.Fatalf("Name = %q, want %q", got.Name, name)
	}
	if !got.JoinedAt.Equal(joinedAt) {
		t.Fatalf("JoinedAt = %v, want %v", got.JoinedAt, joinedAt)
	}
}

// TestCopyToNil 验证空源对象不会触发复制错误并返回空目标。
func TestCopyToNil(t *testing.T) {
	got, err := CopyTo[sourceUser, targetUser](nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got %#v, want nil", got)
	}
}
