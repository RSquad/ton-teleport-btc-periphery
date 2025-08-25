package alerts

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
)

type AlertService struct {
	alertManager *AlertManager
	period       int64
}

func NewAlertService(
	dataSource AlertDataSource,
	alertDispatcher AlertDispatcher,
	cfg *config.ServicesConfig,
) *AlertService {
	alertManager := NewAlertManager(dataSource, alertDispatcher)

	// pegout.signers
	alertManager.RegisterAlert(
		"pegout.signers",
		NewAlertPegoutMaxSignersCount(),
	)

	// pegout.signing
	alertManager.RegisterAlert(
		"pegout.signing",
		NewAlertPegoutSigningDuration(),
	)

	// pegout.restarts
	alertManager.RegisterAlert(
		"pegout.restarts",
		NewAlertPegoutRestarts(),
	)

	return &AlertService{
		alertManager: alertManager,
		period:       int64(cfg.AlertsCheckPeriod),
	}
}

func (service *AlertService) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("AlertService: stopped")
	logger.DefaultLogStartWork("AlertService: starting...")

	ticker := time.NewTicker(time.Duration(service.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("AlertService received shutdown signal...")
			return
		case <-ticker.C:
			service.alertManager.CheckAll()
		}
	}
}
