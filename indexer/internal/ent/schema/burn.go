package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Burn struct {
	ent.Schema
}

func (Burn) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("externalId").
			Unique().
			Immutable().
			Annotations(entgql.MapsTo("externalId")),
		field.Text("senderAddr").
			NotEmpty().
			Immutable().
			Annotations(entgql.MapsTo("senderAddr")),
		field.Text("amount").
			NotEmpty().
			Immutable(),
		field.Text("bitcoinScript").
			NotEmpty().
			Immutable().
			Annotations(entgql.MapsTo("bitcoinScript")),
	}
}

func (Burn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tonMsg", TonMsg.Type).
			Ref("burn").
			Unique().
			Required().
			Annotations(entgql.MapsTo("tonMsg")),
		edge.From("pegout", Pegout.Type).
			Ref("burn").
			Unique().
			Required(),
	}
}

func (Burn) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
	}
}
