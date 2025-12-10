package alerts

import (
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

func TestAlertBtcBlockDelta(t *testing.T) {
	blockHashStr := "00000000000000000000000000000000ffaaff0099887766554433221100ddcc"
	runbookUrl := mutils.CreateHTMLHyperlink("link", "http://runbook/BtcBlockDelta.md")
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
			Expect: TestResWant{Severity: SEVERITY_OK, Description: "OK", Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_OK, Description: "OK", Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_OK, Description: "OK", Err: nil},
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
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("There is a block-height delta of 2 between the BitcoinClient contract (height 100: 95 blocks + 5 confirmations) and the Bitcoin network (height 102).\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("There is a block-height delta of 3 between the BitcoinClient contract (height 100: 95 blocks + 5 confirmations) and the Bitcoin network (height 103).\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("There is a block-height delta of 4 between the BitcoinClient contract (height 100: 95 blocks + 5 confirmations) and the Bitcoin network (height 104).\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("There is a block-height delta of 2 between the BitcoinClient contract (height 102: 97 blocks + 5 confirmations) and the Bitcoin network (height 104).\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("There is a block-height delta of 2 between the BitcoinClient contract (height 102: 97 blocks + 5 confirmations) and the Bitcoin network (height 104).\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
	}

	cfg := config.ServicesConfig{}
	cfg.AlertBtcBlockDeltaHeightCrit = 3
	cfg.AlertBtcBlockDeltaHeightWarn = 2

	DoAlertTests(t, tests, NewAlertBtcBlockDelta(&cfg))
}
