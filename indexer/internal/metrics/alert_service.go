package metrics

import (
	"context"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type AlertService struct{}

func NewAlertService() (*AlertService, error) {
	return &AlertService{}, nil
}

func (s *AlertService) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("AlertService: stopped")
	logger.DefaultLogStartWork("AlertService: starting...")
}
