package ton

import (
    "context"
    "os"

    "github.com/xssnick/tonutils-go/liteclient"
    "github.com/xssnick/tonutils-go/ton"
)

type Client struct {
    Pool *liteclient.ConnectionPool
    API  *ton.APIClient
}

func NewClient() (*Client, error) {
    pool := liteclient.NewConnectionPool()

    configURL := os.Getenv("COMMON_TON_CONFIG_URL")

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
