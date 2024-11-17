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
			Immutable(),
		field.Text("senderAddr").
			NotEmpty().
			Immutable(),
		field.Text("amount").
			NotEmpty().
			Immutable(),
		field.Text("bitcoinTxId").
			Unique().
			NotEmpty().
			Immutable(),
		field.Text("bitcoinScript").
			NotEmpty().
			Immutable(),
	}
}

func (Burn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tonMsg", TonMsg.Type).
			Ref("burn").
			Unique().
			Required(),
	}
}

func (Burn) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
