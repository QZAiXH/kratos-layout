package schema

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// SoftDeleteMixin 为 Ent schema 提供软删除能力。
type SoftDeleteMixin struct {
	mixin.Schema // Schema 是 Ent mixin 基础实现。
}

// Fields 返回软删除字段。
func (SoftDeleteMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("delete_at").
			Optional().
			Nillable().
			Comment("软删除时间"),
	}
}

// softDeleteKey 标记当前上下文是否跳过软删除过滤。
type softDeleteKey struct{}

// SkipSoftDelete 返回跳过软删除过滤和软删除 hook 的上下文。
func SkipSoftDelete(parent context.Context) context.Context {
	return context.WithValue(parent, softDeleteKey{}, true)
}

// Interceptors 返回自动排除软删除记录的查询过滤器。
func (d SoftDeleteMixin) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		ent.TraverseFunc(func(ctx context.Context, q ent.Query) error {
			if skip, _ := ctx.Value(softDeleteKey{}).(bool); skip {
				return nil
			}
			if w, ok := q.(interface{ WhereP(...func(*sql.Selector)) }); ok {
				d.P(w)
			}
			return nil
		}),
	}
}

// P 添加软删除数据过滤条件。
func (d SoftDeleteMixin) P(w interface{ WhereP(...func(*sql.Selector)) }) {
	w.WhereP(sql.FieldIsNull(d.Fields()[0].Descriptor().Name))
}
