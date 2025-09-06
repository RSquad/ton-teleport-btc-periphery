package alerts

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutFeeNotReset(t *testing.T) {
	pegoutAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")

	pegoutLabels1 := Labels{
		"bitcoin_tx_id": "f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b",
		"pegout_addr":   pegoutAddress1.StringRaw(),
	}

	bitcoin_tx_id_1, _ := hex.DecodeString(pegoutLabels1["bitcoin_tx_id"])

	pegoutLabelsEmpty := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
		{
			Name: "SEVERITY_OK (height delta > 0, nextSvb = 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             pegoutAddress1,
						BitcoinTxId:      bitcoin_tx_id_1,
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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (height delta = 0, nextSvb = 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             pegoutAddress1,
						BitcoinTxId:      bitcoin_tx_id_1,
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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (height delta > 0, nextSvb > 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             pegoutAddress1,
						BitcoinTxId:      bitcoin_tx_id_1,
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutFeeNotReset())
}
