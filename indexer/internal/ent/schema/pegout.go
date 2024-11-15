package schema

import "entgo.io/ent"

// Pegout holds the schema definition for the Pegout entity.
type Pegout struct {
	ent.Schema
}

// Fields of the Pegout.
func (Pegout) Fields() []ent.Field {
	return nil
}

// Edges of the Pegout.
func (Pegout) Edges() []ent.Edge {
	return nil
}
