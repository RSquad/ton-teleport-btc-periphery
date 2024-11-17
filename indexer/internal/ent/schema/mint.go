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
		field.Text("receiverAddr").
			NotEmpty().
			Immutable(),
		field.Text("amount").
			NotEmpty().
			Immutable(),
		field.Text("bitcoinTxId").
			Unique().
			NotEmpty().
			Immutable(),
	}
}

func (Mint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tonMsg", TonMsg.Type).
			Ref("mint").
			Unique().
			Required(),
	}
}

func (Mint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
