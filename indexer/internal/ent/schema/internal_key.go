package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type InternalKey struct {
	ent.Schema
}

func (InternalKey) Fields() []ent.Field {
	return []ent.Field{
		field.Text("key").
			NotEmpty().
			Immutable(),
		field.Time("completedAt").
			Immutable().
			Annotations(entgql.MapsTo("completedAt")),
	}
}

func (InternalKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("tonMsg", TonMsg.Type).
			Ref("internalKey").
			Unique().
			Required().
			Annotations(entgql.MapsTo("tonMsg")),
	}
}

func (InternalKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate()),
	}
}
