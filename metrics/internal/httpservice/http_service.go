package httpservice

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

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
		Str("component", "HttpServer").
		Msgf("listening on %s", httpBindAddrStr)

	srv := &http.Server{
		Addr:    httpBindAddrStr,
		Handler: handlerWithCORS,
	}

	var isStopRequested atomic.Bool

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !isStopRequested.Load() {
				logger.Log.Error().
					Str("component", "HttpServer").
					Err(err).
					Msg("terminated")
			}
		}
	}()

	<-ctx.Done()
	isStopRequested.Store(true)
	log.Println("shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v; forcing close", err)
		_ = srv.Close()
	}

	logger.Log.Info().
		Str("component", "HttpServer").
		Msg("stopped")
}
