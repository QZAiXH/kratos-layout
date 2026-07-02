package id

import (
	"strings"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// UUID 生成 UUID 字符串。
func UUID() string {
	return uuid.NewString()
}

// ULID 生成可选前缀的 ULID 字符串。
func ULID(prefix string) string {
	id := ulid.Make().String()
	if strings.TrimSpace(prefix) == "" {
		return id
	}
	return prefix + id
}
