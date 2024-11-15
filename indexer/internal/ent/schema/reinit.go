package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Reinit struct {
	ent.Schema
}

func (Reinit) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("externalId").
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
