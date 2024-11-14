package coordinator_contract

import (
	"context"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
)

type CoordinatorContract struct {
	Address *address.Address
	//sender  *WalletContract
	api *ton.APIClient
	ctx context.Context
}

func NewCoordinatorContract(
	api *ton.APIClient,
	address *address.Address,
	ctx context.Context,
) (*CoordinatorContract, error) {
	return &CoordinatorContract{
		Address: address,
		//sender:  sender,
		api: api,
		ctx: ctx,
	}, nil
}
