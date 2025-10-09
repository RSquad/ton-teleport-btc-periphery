package alerts

import (
	"encoding/hex"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertCpfpLength(t *testing.T) {
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
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 0, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
		{
			Name: "SEVERITY_OK (cpfpLen < 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: []byte("2323233"),
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 0, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 2*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (cpfpLen < 10, time < 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: []byte("2323233"),
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 0, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 3*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (cpfpLen < 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: []byte("2323233"),
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 0, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 5*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (10 <= cppfLen < 20, time < 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: nil,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 15, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 7*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (10 <= cppfLen < 20, time > 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: nil,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 15, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 9*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (cpfpLen >= 20 time < 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: []byte("2323"),
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 20, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 11*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (cpfpLen >= 20, time > 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastSignedPegoutDbFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:             (*data_models.PegoutTonAddr)(pegoutAddress1),
						BitcoinTxId:      bitcoin_tx_id_1,
						BitcoinBlockHash: []byte("2323"),
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 20, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 14*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertCpfpLength())
}
