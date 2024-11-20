package pegoutmanager

import (
	"context"
	"log"
	"sync"
	"time"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/toncenterv3"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
)

type PegoutManager struct {
	ctx               context.Context
	repo              *ent.Client
	teleportContract  *teleportcontract.TeleportContract
	tonClient         *tonclient.TonClient
	tonCenterV3Client *toncenterv3.Client
}

func New(
	ctx context.Context,
	repo *ent.Client,
	tonClient *tonclient.TonClient,
	tonCenterV3Client *toncenterv3.Client,
	teleportContract *teleportcontract.TeleportContract,
) (
	*PegoutManager,
	error,
) {
	pegoutManager := &PegoutManager{
		ctx:               ctx,
		repo:              repo,
		teleportContract:  teleportContract,
		tonClient:         tonClient,
		tonCenterV3Client: tonCenterV3Client,
	}

	return pegoutManager, nil
}

func (c *PegoutManager) Run() {
	for {
		pegouts, err := c.repo.Pegout.Query().
			Where(pegout.StatusNEQ(pegout.StatusCompleted)).
			Limit(128).
			All(c.ctx)
		if err != nil {
			log.Printf("[PegoutManager] failed to get pending pegouts: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		if len(pegouts) == 0 {
			log.Printf("[PegoutManager] no pending pegouts found")
			time.Sleep(3 * time.Second)
			continue
		}

		var wg sync.WaitGroup

		block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			log.Printf("[PegoutManager] failed to get current masterchain info: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		for _, pg := range pegouts {
			wg.Add(1)

			go func(pg *ent.Pegout) {
				defer wg.Done()
				c.processPegout(block, pg)
			}(pg)
		}

		wg.Wait()

		time.Sleep(3 * time.Second)
	}
}

func (c *PegoutManager) processPegout(
	block *ton.BlockIDExt,
	pg *ent.Pegout,
) {
	_, err := c.tonClient.API.GetAccount(c.ctx, block, address.MustParseRawAddr(pg.Addr))
	if err != nil {
		log.Printf("[PegoutManager] failed to get account: %v", err)
		return
	}

	switch pg.Status {
	case pegout.StatusCommitting:
		c.handleCommitingPegout(pg)
	}
}

func (c *PegoutManager) handleCommitingPegout(pg *ent.Pegout) {
}
