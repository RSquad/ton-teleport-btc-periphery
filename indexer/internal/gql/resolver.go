package gql

import (
	"github.com/99designs/gqlgen/graphql"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
)

type Resolver struct {
	repo *ent.Client
}

func NewSchema(client *ent.Client) graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: &Resolver{client},
	})
}
