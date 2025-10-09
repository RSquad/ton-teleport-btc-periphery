package alerts

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

func TestAlertBtcBlockDelta(t *testing.T) {
	blockHashStr := "00000000000000000000000000000000ffaaff0099887766554433221100ddcc"

	labels := Labels{
		"blockHash": blockHashStr,
	}

	blockHash, err := chainhash.NewHashFromStr(blockHashStr)
	if err != nil {
		t.Fatalf("%v", err)
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (delta == 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 100, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 95,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_OK (delta == 1)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 101, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 95,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 2*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_OK (delta == 2, time delta < 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 102, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 95,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 3*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (delta == 2)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 102, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 95,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 4*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (delta == 3)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 103, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 95,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 6*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (delta == 4)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 104, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 95,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 8*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (delta == 2)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 104, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 97,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 10*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (delta == 1, time delta < 2 min)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 104, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 98,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 11*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_OK (delta == 1)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 104, nil
				},
				BitcoinClientContractStorageDbFn: func() (*data_models.BitcoinClientContractStorage, error) {
					return &data_models.BitcoinClientContractStorage{
						LastConfirmedBlockHeight: 98,
						ConfirmationsNeeded:      5,
						LastConfirmedBlockHash:   blockHash,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 12*60
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertBtcBlockDelta())
}
