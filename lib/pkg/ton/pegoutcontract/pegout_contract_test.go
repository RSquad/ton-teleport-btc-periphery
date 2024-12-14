package pegoutcontract

import (
	"context"
	"reflect"
	"testing"

	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
	"github.com/xssnick/tonutils-go/address"
)

func TestInputsOrder(t *testing.T) {
	tonConfigUrl := "https://ton-blockchain.github.io/testnet-global.config.json"
	ctx := context.Background()
	client, err := tonclient.NewTonClient(tonConfigUrl)
	if err != nil {
		t.Errorf("failed to create client: %v", err)
	}
	pegoutAddr := address.MustParseRawAddr("0:56a0a55a87d1e439133a6e1213f974a6957b7e268cf424880d8628b8a142663b")
	pegoutContract := New(
		pegoutAddr,
		client,
		ctx,
	)

	block, err := client.API.CurrentMasterchainInfo(ctx)
	if err != nil {
		t.Errorf("call CurrentMasterchainInfo failed: %e", err)
	}
	parts, err := pegoutContract.GetTxParts(block)
	if err != nil {
		t.Errorf("call GetTxParts failed: %e", err)
	}

	expectedTxids := []string{
		"0031efb7c4d925da0593fec2455c70eb180ecd93e0be1e5e7c5da9cb743ec280",
		"03a9c606a7716e197cd73048b39ede36f7e4d2490f5ec1fac2fa191d3a0d1848",
		"2d5a7dd021ef74292dd54e5bf2ccd1bfa6bfc474b65f03a40f06c64aee61427d",
		"7e53137552a30c7198d4619d5bafc1ab3c9e41cfeb574ce3db03043edc424dc9",
	}

	txids := make([]string, len(*parts.Inputs))
	i := 0
	for id := range *parts.Inputs {
		txids[i] = id
		i++
	}

	if !reflect.DeepEqual(expectedTxids, txids) {
		t.Errorf("slices are not equal:\nexpected: %v\nreceived: %v", expectedTxids, txids)
	}
}
