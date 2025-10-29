package alerts

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertCpfpLength(t *testing.T) {
	pegoutAddress, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	bitcoin_tx_id, _ := hex.DecodeString("f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/PegoutCPFP.md")
	tonUrl := mutils.CreateShortLink("link", "http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78")
	btcUrl := mutils.CreateShortLink("link", "http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (no pegouts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return nil, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 0, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (cpfpLen = 1)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 1, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 2*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (cpfpLen = 10, from cache (time delta < 2 min))",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 10, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 3*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (cppfLen = 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 10, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 4*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The CPFP chain length is 10. Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (cppfLen = 15)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 15, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 6*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The CPFP chain length is 15. Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (cpfpLen = 20, from cache (time delta < 2 min))",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 20, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 7*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The CPFP chain length is 15. Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (cpfpLen = 20)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 20, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 8*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("The CPFP chain length is 20. Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (cpfpLen = 1, from cache (time delta < 2 min))",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 1, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 9*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("The CPFP chain length is 20. Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (cpfpLen = 1)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				LastConfirmedPegoutFn: func() (*data_models.Pegout, error) {
					return &data_models.Pegout{
						Addr:        (*data_models.PegoutTonAddr)(pegoutAddress),
						BitcoinTxId: bitcoin_tx_id,
					}, nil
				},
				BtcGetCpfpLengthFn: func(hash *chainhash.Hash) (int, error) {
					return 1, nil
				},
				NowUnixTsFn: func() int64 {
					return 1234560 + 10*60
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertCpfpLength())
}
