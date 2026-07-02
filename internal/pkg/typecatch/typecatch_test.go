package typecatch

import (
	"testing"
	"time"
)

type sourceUser struct {
	Name     *string
	JoinedAt *time.Time
}

type targetUser struct {
	Name     string
	JoinedAt time.Time
}

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

func TestCopyToNil(t *testing.T) {
	got, err := CopyTo[sourceUser, targetUser](nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got %#v, want nil", got)
	}
}
