package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Burn struct {
	ent.Schema
}

func (Burn) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("externalId").
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
		field.Text("tonMsgHash").
			Unique().
			NotEmpty().
			Immutable(),
		field.Time("createdAt").
			Immutable(),
	}
}

func (Burn) Edges() []ent.Edge {
	return nil
}
