package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Reinit struct {
	ent.Schema
}

func (Reinit) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("externalId").
			Unique().
			Immutable(),
		field.Text("amount").
			NotEmpty().
			Immutable(),
		field.Text("bitcoinTxId").
			Unique().
			NotEmpty().
			Immutable(),
		field.Text("bitcoinScript").
			Immutable(),
		field.Text("tonMsgHash").
			Unique().
			NotEmpty().
			Immutable(),
		field.Time("createdAt").
			Immutable(),
	}
}

func (Reinit) Edges() []ent.Edge {
	return nil
}

func (Reinit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
