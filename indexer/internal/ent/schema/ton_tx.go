package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type TonTx struct {
	ent.Schema
}

func (TonTx) Fields() []ent.Field {
	return []ent.Field{
		field.Text("hash").
			Unique().
			NotEmpty().
			Immutable(),
		field.Time("created_at").
			Immutable(),
	}
}

func (TonTx) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("mint", Mint.Type).
			Unique(),
		edge.To("burn", Burn.Type).
			Unique(),
		edge.To("reinit", Reinit.Type).
			Unique(),
		edge.To("internal_key", InternalKey.Type).
			Unique(),
	}
}

func (TonTx) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
	}
}
