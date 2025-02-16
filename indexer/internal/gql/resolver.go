package gql

import (
	"github.com/99designs/gqlgen/graphql"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

type Resolver struct {
	repo             *ent.Client
	bitcoinClient    *bitcoin.Client
	teleportContract *teleportcontract.TeleportContract
}

func NewSchema(
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	teleportContract *teleportcontract.TeleportContract,
) graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: &Resolver{repo, bitcoinClient, teleportContract},
	})
}
