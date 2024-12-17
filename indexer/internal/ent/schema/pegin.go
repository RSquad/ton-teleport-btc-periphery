package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type Pegin struct {
	ent.Schema
}

func (Pegin) Fields() []ent.Field {
	return []ent.Field{
		field.Text("receiverAddr").
			NotEmpty().
			Immutable().
			Annotations(entgql.MapsTo("receiverAddr")),
		field.Text("bitcoinTxId").
			Unique().
			NotEmpty().
			Immutable().
			Annotations(entgql.MapsTo("bitcoinTxId")),
		field.Text("internalKey").
			Optional().
			Immutable().
			Annotations(
				entgql.MapsTo("internalKey"),
			),
		field.Text("recoveryKey").
			Optional().
			Immutable().
			Annotations(
				entgql.MapsTo("recoveryKey"),
			),
	}
}

func (Pegin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("mint", Mint.Type).
			Ref("pegin").
			Unique().
			Required().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput)),
	}
}

func (Pegin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
