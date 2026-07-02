package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// TimeMixin 为 Ent schema 提供版本和时间戳字段。
type TimeMixin struct {
	mixin.Schema // Schema 是 Ent mixin 基础实现。
}

// Fields 返回通用时间字段。
func (TimeMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("version").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("乐观锁版本或最后修改时间"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
	}
}
