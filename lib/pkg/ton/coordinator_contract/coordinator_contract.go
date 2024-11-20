package coordinatorcontract

import (
	"context"

	"github.com/xssnick/tonutils-go/address"

	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
)

type CoordinatorContract struct {
	Addr      *address.Address
	tonClient *tonclient.TonClient
	ctx       context.Context
}

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) *CoordinatorContract {
	return &CoordinatorContract{
		Addr:      addr,
		tonClient: tonClient,
		ctx:       ctx,
	}
}
