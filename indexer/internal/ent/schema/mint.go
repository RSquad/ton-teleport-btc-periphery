package schema

import (
	"entgo.io/ent"
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
		field.Text("tonMsgHash").
			Unique().
			NotEmpty().
			Immutable(),
		field.Time("createdAt").
			Immutable(),
	}
}

func (Mint) Edges() []ent.Edge {
	return nil
}
