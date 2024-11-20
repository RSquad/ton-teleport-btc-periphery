package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Reinit struct {
	ent.Schema
}

func (Reinit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("externalID").
			Unique().
			Immutable(),
		field.Text("amount").
			NotEmpty().
			Immutable(),
		field.Text("bitcoinTxId").
			NotEmpty().
			Immutable(),
		field.Text("bitcoinScript").
			Immutable(),
	}
}

func (Reinit) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tonMsg", TonMsg.Type).
			Ref("reinit").
			Unique().
			Required(),
		edge.From("pegout", Pegout.Type).
			Ref("reinit").
			Unique().
			Required(),
	}
}

func (Reinit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
