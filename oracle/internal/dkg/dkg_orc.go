package dkg

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/xssnick/tonutils-go/address"
)

type Orc struct {
	Runner *Runner
}

func NewOrc(
	tonClient *tonclient.TonClient,
	cfg *cfg.Cfg,
) *Orc {
	coordinatorContract := coordinatorcontract.New(
		signer.New(cfg.Secret),
		address.MustParseAddr(cfg.CoordinatorContractAddr),
		tonClient,
		context.Background(),
	)
	fetcher := NewFetcher(coordinatorContract)
	executor := NewExecutor()
	runner := NewRunner(fetcher, executor, 3*time.Second)

	return &Orc{
		Runner: runner,
	}
}
