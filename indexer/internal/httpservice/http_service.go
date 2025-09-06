package httpservice

import (
	"context"
	"net/http"

	"entgo.io/contrib/entgql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/rs/cors"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/gql"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type HttpService struct {
	repo             *ent.Client
	bitcoinClient    *bitcoin.Client
	tonClient        *tonclient.TonClient
	teleportContract *teleportcontract.TeleportContract
}

func New(
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	teleportContract *teleportcontract.TeleportContract,
) *HttpService {
	return &HttpService{
		repo:             repo,
		bitcoinClient:    bitcoinClient,
		tonClient:        tonClient,
		teleportContract: teleportContract,
	}
}

func (s *HttpService) Work(ctx context.Context) {
	srv := handler.NewDefaultServer(
		gql.NewSchema(s.repo, s.bitcoinClient, s.teleportContract, s.tonClient),
	)
	srv.Use(entgql.Transactioner{TxOpener: s.repo})

	mux := http.NewServeMux()
	mux.Handle("/indexer/graphql", srv)
	mux.Handle("/", playground.ApolloSandboxHandler("Indexer", "/indexer/graphql"))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handlerWithCORS := c.Handler(mux)

	logger.Log.Info().
		Str("component", "main").
		Msg("listening on :3000")
	if err := http.ListenAndServe(":3000", handlerWithCORS); err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("http server terminated")
	}
}
