package alerts

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutSigningDuration(t *testing.T) {
	signingTimeout := int64(60 * 22) // 22 min
	beginTs := int64(123456)
	pegoutAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	pegoutAddress2, _ := address.ParseAddr("Ef8VjV6LGTyiNLzefOm1dpuCMLcoewhqfQubtgbWcPwt2Gwp")

	bitcoin_tx_id_1, _ := hex.DecodeString("f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	bitcoin_tx_id_2, _ := hex.DecodeString("3d46303861d5336c3ebdea3a20883a1cb77f4f3a66a2fb5e6494d3a0ab878bd1")

	tonUrl := mutils.CreateShortLink("link", "http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78")
	btcUrl := mutils.CreateShortLink("link", "http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/PegoutSigningDuration.md")

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (new unsigned pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(0, 0),
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						SigningTimeout: uint32(signingTimeout),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (same unsigned pegout, 1 minute later)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*1
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (same unsigned pegout, 11 minute later)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*11
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (same unsigned pegout, 22 minute later)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*22
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Pegout transaction was not signed within 22 minutes.\nPegout: %s.\nBitcoin TX: %s.\nRunbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (same unsigned pegout, 2 minute later after restart, 22 minutes total)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*2, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*22
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Pegout transaction was not signed within 22 minutes.\nPegout: %s.\nBitcoin TX: %s.\nRunbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (new unsigned pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					beginTs = beginTs + 60*24 // Update beginTs for new pegout pegoutAddress2
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress2,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_2,
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*25
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK, all pegout are signed",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return nil, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutSigningDuration())
}
