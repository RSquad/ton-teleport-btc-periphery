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
		field.Text("receiver_addr").
			NotEmpty().
			Immutable(),
		field.Text("bitcoin_tx_id").
			Unique().
			NotEmpty().
			Immutable(),
		field.Int("vout_index").
			Optional().
			Immutable().
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput)),
		field.Text("internal_key").
			Optional().
			Immutable(),
		field.Text("recovery_key").
			Optional().
			Immutable(),
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
