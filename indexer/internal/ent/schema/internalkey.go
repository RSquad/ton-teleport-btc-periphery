package schema

import "entgo.io/ent"

// InternalKey holds the schema definition for the InternalKey entity.
type InternalKey struct {
	ent.Schema
}

// Fields of the InternalKey.
func (InternalKey) Fields() []ent.Field {
	return nil
}

// Edges of the InternalKey.
func (InternalKey) Edges() []ent.Edge {
	return nil
}
