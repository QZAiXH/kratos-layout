package schema

import (
	"context"

	"github.com/QZAiXH/kratos-layout/internal/pkg/authz"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// ByUserMixin 为 Ent schema 提供创建人和更新人字段。
type ByUserMixin struct {
	mixin.Schema // Schema 是 Ent mixin 基础实现。
}

// Fields 返回创建人和更新人字段。
func (ByUserMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("created_by").
			SchemaType(map[string]string{dialect.Postgres: "char(26)"}).
			Comment("创建人用户 ULID"),
		field.String("updated_by").
			SchemaType(map[string]string{dialect.Postgres: "char(26)"}).
			Comment("最后更新人用户 ULID"),
	}
}

// Hooks 返回自动写入当前用户 ID 的 hook。
func (ByUserMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		setUser,
	}
}

// setUser 根据当前上下文写入创建人或更新人。
func setUser(next ent.Mutator) ent.Mutator {
	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		switch m.Op() {
		case ent.OpUpdateOne, ent.OpUpdate:
			userID, err := authz.GetUserIDFromContext(ctx)
			if err != nil {
				return next.Mutate(ctx, m)
			}
			if err := m.SetField("updated_by", userID); err != nil {
				return nil, err
			}
		case ent.OpCreate:
			userID, err := authz.GetUserIDFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if err := m.SetField("created_by", userID); err != nil {
				return nil, err
			}
			if err := m.SetField("updated_by", userID); err != nil {
				return nil, err
			}
		}
		return next.Mutate(ctx, m)
	})
}

// Edges 返回用户审计 mixin 的边定义。
func (ByUserMixin) Edges() []ent.Edge {
	return nil
}
