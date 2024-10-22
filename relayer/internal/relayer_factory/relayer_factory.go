package relayerfactory

import (
    "context"
    "fmt"
    "os"

    "github.com/xssnick/tonutils-go/address"

    "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
    "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
    "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/block_relayer"
    "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/pegout_relayer"
)

type Relayer interface {
    Relay() error
}

type RelayerFactory struct {
    bitcoinClient *bitcoin.Client
    tonClient     *ton.Client
}

func NewRelayerFactory(bitcoinClient *bitcoin.Client, tonClient *ton.Client) *RelayerFactory {
    return &RelayerFactory{
        bitcoinClient: bitcoinClient,
        tonClient:     tonClient,
    }
}

func (c *RelayerFactory) CreateRelayer(relayerType string, sender *ton.WalletContract) (Relayer, error) {
    switch relayerType {
    case "block":
        bitcoinClientContract := ton.NewBitcoinClientContract(
            c.tonClient.API,
            address.MustParseAddr(os.Getenv("COMMON_TON_CONTRACT_BITCLIENT_ADDR")),
            sender,
            context.Background(),
        )
        return blockrelayer.NewBlockRelayer(c.bitcoinClient, bitcoinClientContract)
    case "pegout":
        teleportContract := ton.NewTeleportContract(
            c.tonClient.API,
            address.MustParseAddr(os.Getenv("COMMON_TON_CONTRACT_TELEPORT_ADDR")),
            sender,
            context.Background(),
        )
        return pegoutrelayer.NewPegoutRelayer(c.bitcoinClient, teleportContract)
    default:
        return nil, fmt.Errorf("unknown relayer type: %s", relayerType)
    }
}
