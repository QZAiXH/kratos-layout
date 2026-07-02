package id

import (
	"strings"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

func UUID() string {
	return uuid.NewString()
}

func ULID(prefix string) string {
	id := ulid.Make().String()
	if strings.TrimSpace(prefix) == "" {
		return id
	}
	return prefix + id
}
