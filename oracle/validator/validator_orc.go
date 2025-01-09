package validator

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
)

type Orc struct{}

func NewOrc(
	tonClient *tonclient.TonClient,
	cfg *cfg.Cfg,
) *Orc {
	standaloneMode := cfg.StandaloneMode

	return &Orc{}
}
