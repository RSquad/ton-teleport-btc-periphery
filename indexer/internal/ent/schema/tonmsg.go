package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type TonMsg struct {
	ent.Schema
}

func (TonMsg) Fields() []ent.Field {
	return []ent.Field{
		field.Text("hash").
			Unique().
			NotEmpty().
			Immutable(),
		field.Time("createdAt").
			Immutable(),
	}
}

func (TonMsg) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("mint", Mint.Type).
			Unique(),
		edge.To("burn", Burn.Type).
			Unique(),
		edge.To("reinit", Reinit.Type).
			Unique(),
	}
}

func (TonMsg) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
