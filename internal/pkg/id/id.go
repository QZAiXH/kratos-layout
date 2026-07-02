package id

import (
	"strings"

	"github.com/oklog/ulid/v2"
)

// ULID 生成可选前缀的 ULID 字符串。
func ULID(prefix string) string {
	id := ulid.Make().String()
	if strings.TrimSpace(prefix) == "" {
		return id
	}
	return prefix + id
}
