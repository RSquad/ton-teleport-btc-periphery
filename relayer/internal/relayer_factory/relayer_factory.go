package relayerfactory

import (
	"context"
	"fmt"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	bitcoinclientcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	blockrelayer "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/block_relayer"
	pegoutrelayer "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/pegout_relayer"
)

type Relayer interface {
	Relay() error
}

type RelayerFactory struct {
	bitcoinClient *bitcoin.Client
	tonClient     *tonclient.TonClient
}

func NewRelayerFactory(bitcoinClient *bitcoin.Client, tonClient *tonclient.TonClient) *RelayerFactory {
	return &RelayerFactory{
		bitcoinClient: bitcoinClient,
		tonClient:     tonClient,
	}
}

func (c *RelayerFactory) CreateRelayer(
	relayerType string,
	sender *jwv4r2contract.JWV4R2Contract,
	bitcoinClientContractAddrStr string,
	teleportContractAddrStr string,
) (
	Relayer,
	error,
) {
	bitcoinClientContractAddr, err := address.ParseAddr(bitcoinClientContractAddrStr)
	if err != nil {
		return nil, fmt.Errorf("parsing the Bitcoin Client Contract address '%s' failed", bitcoinClientContractAddrStr)
	}

	bitcoinClientContract := bitcoinclientcontract.NewBitcoinClientContract(
		bitcoinClientContractAddr,
		c.tonClient,
		sender,
		context.Background(),
	)
	switch relayerType {
	case "block":
		return blockrelayer.NewBlockRelayer(c.bitcoinClient, bitcoinClientContract)
	case "pegout":
		teleportContractAddr, err := address.ParseAddr(teleportContractAddrStr)
		if err != nil {
			return nil, fmt.Errorf("parsing the Teleport Contract address '%s' failed", teleportContractAddrStr)
		}

		teleportContract := teleportcontract.New(
			teleportContractAddr,
			c.tonClient,
			sender,
			nil,
			context.Background(),
		)
		return pegoutrelayer.NewPegoutRelayer(c.bitcoinClient, teleportContract, bitcoinClientContract)
	default:
		return nil, fmt.Errorf("[RelayerFactory] unknown relayer type: %s", relayerType)
	}
}
