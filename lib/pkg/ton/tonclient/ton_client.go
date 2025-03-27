package tonclient

import (
	"context"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type TonClient struct {
	Pool *liteclient.ConnectionPool
	API  *ton.APIClient
}

func New(configPathOrURL string) (*TonClient, error) {
	pool := liteclient.NewConnectionPool()
	if strings.HasPrefix(configPathOrURL, "http://") || strings.HasPrefix(configPathOrURL, "https://") {
		err := pool.AddConnectionsFromConfigUrl(context.Background(), configPathOrURL)
		if err != nil {
			return nil, err
		}
	} else {
		err := pool.AddConnectionsFromConfigFile(configPathOrURL)
		if err != nil {
			return nil, err
		}
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

func (tc *TonClient) GetBalance(addr *address.Address) (tlb.Coins, error) {
	account, err := tc.FetchAcc(addr, nil)
	if err != nil {
		return tlb.Coins{}, err
	}

	return account.State.Balance, nil
}
