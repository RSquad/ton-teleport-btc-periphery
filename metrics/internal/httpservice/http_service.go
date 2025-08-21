package httpservice

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/rs/cors"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type HttpService struct {
	bitcoinClient    *bitcoin.Client
	tonClient        *tonclient.TonClient
	teleportContract *teleportcontract.TeleportContract
	db               *sql.DB
}

func New(
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	teleportContract *teleportcontract.TeleportContract,
	db *sql.DB,
) *HttpService {
	return &HttpService{
		bitcoinClient:    bitcoinClient,
		tonClient:        tonClient,
		teleportContract: teleportContract,
		db:               db,
	}
}

func (s *HttpService) Work(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics/api", NewJsonApiHandler(s.db, s.tonClient))

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
