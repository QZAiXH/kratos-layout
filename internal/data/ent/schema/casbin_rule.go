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

// Fields 定义 Casbin 策略字段。
func (CasbinRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("ptype").Default(""),
		field.String("v0").Default(""),
		field.String("v1").Default(""),
		field.String("v2").Default(""),
		field.String("v3").Default(""),
		field.String("v4").Default(""),
		field.String("v5").Default(""),
	}
}
