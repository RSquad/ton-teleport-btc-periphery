package gql

import (
	"github.com/99designs/gqlgen/graphql"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/peginmanager"
)

type Resolver struct {
	repo         *ent.Client
	peginManager *peginmanager.PeginManager
}

func NewSchema(client *ent.Client, peginManager *peginmanager.PeginManager) graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: &Resolver{client, peginManager},
	})
}
