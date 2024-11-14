package ton

import (
	"context"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

type Client struct {
	Pool *liteclient.ConnectionPool
	API  *ton.APIClient
}

func NewClient(configURL string) (*Client, error) {
	pool := liteclient.NewConnectionPool()

	err := pool.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		return nil, err
	}

	api := ton.NewAPIClient(pool)

	return &Client{
		Pool: pool,
		API:  api,
	}, nil

}
