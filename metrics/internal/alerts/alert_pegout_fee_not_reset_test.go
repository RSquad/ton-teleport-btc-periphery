package alerts

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutFeeNotReset(t *testing.T) {
	pegoutAddress, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	bitcoin_tx_id, _ := hex.DecodeString("f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")

	tonUrl := mutils.CreateShortLink("link", "http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78")
	btcUrl := mutils.CreateShortLink("link", "http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/PegoutFeeNotReset.md")

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (height delta > 0, nextSvb = 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId:      bitcoin_tx_id,
						BitcoinBlockHash: []byte("2323233"),
					}, nil
				},
				BtcGetBlockHeightByHashFn: func(hash *chainhash.Hash) (int64, error) {
					return 1234567, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 1234567 + 1,
					}, nil
				},
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						NextSVB: 0,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (height delta = 0, nextSvb = 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId:      bitcoin_tx_id,
						BitcoinBlockHash: nil,
					}, nil
				},
				BtcGetBlockHeightByHashFn: func(hash *chainhash.Hash) (int64, error) {
					return 1234567, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 1234567,
					}, nil
				},
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						NextSVB: 0,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (height delta > 0, nextSvb > 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId:      bitcoin_tx_id,
						BitcoinBlockHash: []byte("2323"),
					}, nil
				},
				BtcGetBlockHeightByHashFn: func(hash *chainhash.Hash) (int64, error) {
					return 1234567, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 1234567 + 1,
					}, nil
				},
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						NextSVB: 123,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Pegout transaction already has 1 (pegoutBlockHeight 1234567, lastConfirmedBlockHeight 1234568) confirmations but the fee has not been reset (nextSvb = 123).\nPegout: %s.\nBitcoin TX: %s.\nRunbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutFeeNotReset())
}
