package httpservice

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/alerts"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/metrics"
)

type HttpService struct {
	metricsManager      *metrics.MetricsManager
	alertManager        *alerts.AlertManager
	httpPort            int
	alertsTestApiEnable bool
}

func New(
	metricsManager *metrics.MetricsManager,
	alertManager *alerts.AlertManager,
	cfg *config.ServicesConfig,
) *HttpService {
	return &HttpService{
		metricsManager:      metricsManager,
		alertManager:        alertManager,
		httpPort:            cfg.HttpPort,
		alertsTestApiEnable: cfg.AlertsTestApiEnable,
	}
}

func (s *HttpService) Work(ctx context.Context) {
	mux := http.NewServeMux()
	mux.Handle("/metrics/api", NewJsonApiHandler(s.metricsManager, s.alertManager))
	mux.Handle("/metrics/prom", promhttp.Handler())

	if s.alertsTestApiEnable {
		mux.Handle("/metrics/alerts_testing", NewAlertsTestingApiHandler(s.alertManager))
	}

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
