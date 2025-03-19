package tonclient

import (
	"context"
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type TonClient struct {
	Pool *liteclient.ConnectionPool
	API  *ton.APIClient
}

func New(configURL string) (*TonClient, error) {
	pool := liteclient.NewConnectionPool()

	err := pool.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		return nil, err
	}

	api := ton.NewAPIClient(pool).WithRetry(5)

	return &TonClient{
		Pool: pool,
		API:  api.(*ton.APIClient),
	}, nil
}

func (tc *TonClient) FetchAcc(
	addr *address.Address,
	block *ton.BlockIDExt,
) (*tlb.Account, error) {
	var err error
	if block == nil {
		block, err = tc.API.CurrentMasterchainInfo(context.Background())
		if err != nil {
			return nil, err
		}
	}

	return tc.API.GetAccount(context.Background(), block, addr)
}

func (tc *TonClient) GetBalance(addr *address.Address) (*big.Int, error) {

	account, err := tc.FetchAcc(addr, nil)
	if err != nil {
		return nil, err
	}

	return account.State.Balance.Nano(), nil
}
