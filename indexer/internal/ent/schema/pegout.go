package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Pegout struct {
	ent.Schema
}

func (Pegout) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("externalID").
			Unique().
			Immutable(),
		field.Text("addr").
			NotEmpty().
			Immutable(),
		field.Enum("status").
			NamedValues(
				"Signing", "SIGNING",
				"Completed", "COMPLETED",
			).
			Default("SIGNING"),
		field.Text("bitcoinTxRaw").
			Default(""),
		field.Text("bitcoinTxId").
			Default(""),
		field.Bool("isBitcoinTxSent").
			Default(false),
	}
}

func (Pegout) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("burn", Burn.Type).
			Unique(),
		edge.To("reinit", Reinit.Type).
			Unique(),
	}
}

func (Pegout) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
