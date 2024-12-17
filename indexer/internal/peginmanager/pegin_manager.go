package peginmanager

import (
	"context"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/toncenterv3"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type PeginManager struct {
	ctx               context.Context
	repo              *ent.Client
	bitcoinClient     *bitcoin.Client
	tonClient         *tonclient.TonClient
	tonCenterV3Client *toncenterv3.Client
	teleportContract  *teleportcontract.TeleportContract
}

func New(
	ctx context.Context,
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	tonCenterV3Client *toncenterv3.Client,
	teleportContract *teleportcontract.TeleportContract,
) *PeginManager {
	return &PeginManager{
		ctx:               ctx,
		repo:              repo,
		bitcoinClient:     bitcoinClient,
		tonClient:         tonClient,
		tonCenterV3Client: tonCenterV3Client,
		teleportContract:  teleportContract,
	}
}
