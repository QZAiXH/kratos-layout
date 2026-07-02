package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// CasbinRule 定义 Casbin 策略表结构。
type CasbinRule struct {
	ent.Schema // Schema 嵌入 Ent 基础 schema。
}

// Annotations 指定 Casbin 策略表名。
func (CasbinRule) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "casbin_rule"}}
}

// Mixin 返回 Casbin 策略表通用字段。
func (CasbinRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		StringIDMixin{},
		TimeMixin{},
	}
}

// Fields 定义 Casbin 策略字段。
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("ptype").Default("").Comment("Casbin 策略类型"),
		field.String("v0").Default("").Comment("Casbin 策略字段 0"),
		field.String("v1").Default("").Comment("Casbin 策略字段 1"),
		field.String("v2").Default("").Comment("Casbin 策略字段 2"),
		field.String("v3").Default("").Comment("Casbin 策略字段 3"),
		field.String("v4").Default("").Comment("Casbin 策略字段 4"),
		field.String("v5").Default("").Comment("Casbin 策略字段 5"),
	}
}
