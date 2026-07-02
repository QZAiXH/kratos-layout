package schema

import (
	"strings"

	"github.com/QZAiXH/kratos-layout/internal/pkg/id"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

const ulidLength = 26

// StringIDMixin 为 Ent schema 提供字符串 ULID 主键。
type StringIDMixin struct {
	mixin.Schema        // Schema 是 Ent mixin 基础实现。
	Prefix       string // Prefix 是可选业务前缀。
}

// idMaxLength 返回带前缀 ID 的最大长度。
func (m StringIDMixin) idMaxLength() int {
	return len(strings.TrimSpace(m.Prefix)) + ulidLength
}

// Fields 返回 ULID 主键字段。
func (m StringIDMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			MaxLen(m.idMaxLength()).
			Immutable().
			Unique().
			DefaultFunc(func() string {
				return id.ULID(strings.TrimSpace(m.Prefix))
			}).
			Comment("ULID 主键"),
	}
}
