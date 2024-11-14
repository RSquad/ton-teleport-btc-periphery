package ton

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/config"
)

type TonService struct {
	config        config.OracleConfig
	tonClient     *ton.Client
	tcCoordinator *ton.CoordinatorContract
}

func NewTonService(config config.OracleConfig, tonClient *ton.Client, tcCoordinator *ton.CoordinatorContract) *TonService {
	return &TonService{
		config:        config,
		tonClient:     tonClient,
		tcCoordinator: tcCoordinator,
	}
}
