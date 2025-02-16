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
		field.Int64("external_id").
			Unique().
			Immutable(),
		field.Text("sender_addr").
			NotEmpty().
			Immutable(),
		field.Text("amount").
			NotEmpty().
			Immutable(),
		field.Text("bitcoin_script").
			NotEmpty().
			Immutable(),
	}
}

func (Burn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ton_tx", TonTx.Type).
			Ref("burn").
			Unique().
			Required(),
		edge.From("pegout", Pegout.Type).
			Ref("burn").
			Unique().
			Required(),
	}
}

func (Burn) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
	}
}
