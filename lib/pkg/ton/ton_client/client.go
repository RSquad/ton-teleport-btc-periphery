package tonclient

import (
	"context"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

type TonClient struct {
	Pool *liteclient.ConnectionPool
	API  *ton.APIClient
}

func NewTonClient(configURL string) (*TonClient, error) {
	pool := liteclient.NewConnectionPool()

	err := pool.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		return nil, err
	}

	api := ton.NewAPIClient(pool)

	return &TonClient{
		Pool: pool,
		API:  api,
	}, nil

}
