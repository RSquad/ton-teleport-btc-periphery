package alerts

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutInMempool(t *testing.T) {
	beginTs := int64(123456)

	pegoutAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	pegoutAddress2, _ := address.ParseAddr("Ef8VjV6LGTyiNLzefOm1dpuCMLcoewhqfQubtgbWcPwt2Gwp")

	pegoutLabels1 := Labels{
		"bitcoin_tx_id": "f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b",
		"pegout_addr":   pegoutAddress1.StringRaw(),
	}

	pegoutLabels2 := Labels{
		"bitcoin_tx_id": "3d46303861d5336c3ebdea3a20883a1cb77f4f3a66a2fb5e6494d3a0ab878bd1",
		"pegout_addr":   pegoutAddress2.StringRaw(),
	}

	bitcoin_tx_id_1, _ := hex.DecodeString(pegoutLabels1["bitcoin_tx_id"])
	bitcoin_tx_id_2, _ := hex.DecodeString(pegoutLabels2["bitcoin_tx_id"])

	pegouts := []*data_models.Pegout{
		{
			Id:          1,
			Addr:        pegoutAddress1,
			BitcoinTxId: bitcoin_tx_id_1,
		},
		{
			Id:          2,
			Addr:        pegoutAddress2,
			BitcoinTxId: bitcoin_tx_id_2,
		},
	}

	pegoutLabelsEmpty := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (no pegouts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
		{
			Name: "SEVERITY_OK (pegout1 not in mempool or block, duration: 0 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (pegout1 not in mempool or block, duration: 1 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 1*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (pegout1 not in mempool or block, duration: 10 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 10*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (pegout1 in mempool, next pegout2 (will be checked next time), duration: 0 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return &btcjson.GetMempoolEntryResult{}, nil
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (pegout2 not in mempool or block duration: 0 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels2, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (pegout2 not in mempool or block duration: 10 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 10*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: pegoutLabels2, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (pegout2 not in mempool or block duration: 10 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return nil, fmt.Errorf("Not found")
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 40*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels2, Err: nil},
		},
		{
			Name: "SEVERITY_OK (pegout2 in block, no new pegouts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				BtcGetMempoolEntryFn: func(txHash string) (*btcjson.GetMempoolEntryResult, error) {
					return nil, fmt.Errorf("Not found")
				},
				BtcGetBlockHashByTxIdFn: func(txID *chainhash.Hash) (*chainhash.Hash, error) {
					return &chainhash.Hash{}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 44*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels2, Err: nil},
		},
		{
			Name: "SEVERITY_OK (no new pegouts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutsDbFn: func(limit uint) ([]*data_models.Pegout, error) {
					return pegouts, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 46*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutInMempool())
}
