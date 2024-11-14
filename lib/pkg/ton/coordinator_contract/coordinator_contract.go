package coordinatorcontract

import (
	"context"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
)

type CoordinatorContract struct {
	Address   *address.Address
	tonClient *tonclient.TonClient
	ctx       context.Context
}

func NewCoordinatorContract(
	address *address.Address,
	tonClient *tonclient.TonClient,
	ctx context.Context,
) (*CoordinatorContract, error) {
	return &CoordinatorContract{
		Address:   address,
		tonClient: tonClient,
		ctx:       ctx,
	}, nil
}
