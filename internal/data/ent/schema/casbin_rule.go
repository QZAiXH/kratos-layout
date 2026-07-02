package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type CasbinRule struct {
	ent.Schema
}

func (CasbinRule) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "casbin_rule"}}
}

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
