package httpservice

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
)

type HttpService struct {
	bitcoinClient    *bitcoin.Client
	tonClient        *tonclient.TonClient
	teleportContract *teleportcontract.TeleportContract
	db               *sql.DB
	httpPort         int
}

func New(
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	teleportContract *teleportcontract.TeleportContract,
	db *sql.DB,
	cfg *config.ServicesConfig,
) *HttpService {
	return &HttpService{
		bitcoinClient:    bitcoinClient,
		tonClient:        tonClient,
		teleportContract: teleportContract,
		db:               db,
		httpPort:         cfg.HttpPort,
	}
}

func (s *HttpService) Work(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics/api", NewJsonApiHandler(s.db, s.tonClient))
	mux.Handle("/metrics/prom", promhttp.Handler())

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handlerWithCORS := c.Handler(mux)
	httpBindAddrStr := fmt.Sprintf(":%d", s.httpPort)

	logger.Log.Info().
		Str("component", "main").
		Msgf("listening on %s", httpBindAddrStr)
	if err := http.ListenAndServe(httpBindAddrStr, handlerWithCORS); err != nil {
		logger.Log.Error().
			Str("component", "main").
			Err(err).
			Msg("http server terminated")
	}
}
