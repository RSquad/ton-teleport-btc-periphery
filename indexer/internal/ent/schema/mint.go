package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Mint struct {
	ent.Schema
}

func (Mint) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			NamedValues(
				"Pending", "PENDING",
				"Success", "SUCCESS",
				"Refund", "REFUND",
				"Refunded", "REFUNDED",
			).
			Default("PENDING"),
		field.Text("amount").
			Optional().
			Immutable(),
		field.Time("created_at").
			Immutable(),
	}
}

func (Mint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ton_msg", TonMsg.Type).
			Ref("mint").
			Unique(),
		edge.To("pegin", Pegin.Type).
			Unique(),
	}
}

func (Mint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
	}
}
