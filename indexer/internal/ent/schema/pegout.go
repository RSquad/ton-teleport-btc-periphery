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
		field.Int64("externalId").
			Unique().
			Immutable().
			Annotations(entgql.MapsTo("externalId")),
		field.Text("addr").
			NotEmpty().
			Immutable(),
		field.Enum("status").
			NamedValues(
				"Signing", "SIGNING",
				"Signed", "SIGNED",
				"Confirmed", "CONFIRMED",
			).
			Default("SIGNING"),
		field.Text("bitcoinTxRaw").
			Default("").
			Annotations(entgql.MapsTo("bitcoinTxRaw")),
		field.Text("bitcoinTxId").
			Default("").
			Annotations(entgql.MapsTo("bitcoinTxId")),
		field.Text("bitcoinBlockHash").
			Default("").
			Annotations(entgql.MapsTo("bitcoinBlockHash")),
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
	}
}
