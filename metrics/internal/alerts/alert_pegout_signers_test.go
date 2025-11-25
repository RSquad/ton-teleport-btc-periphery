package alerts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutSigners(t *testing.T) {
	pegoutAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	pegoutAddress2, _ := address.ParseAddr("Ef8VjV6LGTyiNLzefOm1dpuCMLcoewhqfQubtgbWcPwt2Gwp")

	bitcoin_tx_id_1, _ := hex.DecodeString("f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	bitcoin_tx_id_2, _ := hex.DecodeString("3d46303861d5336c3ebdea3a20883a1cb77f4f3a66a2fb5e6494d3a0ab878bd1")

	tonUrl1 := mutils.CreateHTMLHyperlink("link", "http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78")
	tonUrl2 := mutils.CreateHTMLHyperlink("link", "http://ton/-1:158d5e8b193ca234bcde7ce9b5769b8230b7287b086a7d0b9bb606d670fc2dd8")
	btcUrl1 := mutils.CreateHTMLHyperlink("link", "http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	btcUrl2 := mutils.CreateHTMLHyperlink("link", "http://btc/3d46303861d5336c3ebdea3a20883a1cb77f4f3a66a2fb5e6494d3a0ab878bd1")
	runbookUrl := mutils.CreateHTMLHyperlink("link", "http://runbook/PegoutSigners.md")

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (10 of 10 [100%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						SigningMask:   new(big.Int).SetUint64(0b1111111111),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
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
			Name: "SEVERITY_INFO (9 of 10 [90%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						SigningMask:   new(big.Int).SetUint64(0b0111111111),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_INFO,
				Description: Description(fmt.Sprintf("Number of validators allowed to sign pegout is 9 of 10 (90%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s", tonUrl1, btcUrl1, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (8 of 10 [80%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						SigningMask:   new(big.Int).SetUint64(0b0011111111),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("Number of validators allowed to sign pegout is 8 of 10 (80%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s", tonUrl1, btcUrl1, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (7 of 10 [70%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						SigningMask:   new(big.Int).SetUint64(0b0001111111),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Number of validators allowed to sign pegout is 7 of 10 (70%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s", tonUrl1, btcUrl1, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (6 of 10 [60%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						SigningMask:   new(big.Int).SetUint64(0b0001111101),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Number of validators allowed to sign pegout is 6 of 10 (60%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s", tonUrl1, btcUrl1, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (6 of 10 [60%]), next pegout",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress2,
						SigningMask:   new(big.Int).SetUint64(0b0001111101),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_2,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Number of validators allowed to sign pegout is 6 of 10 (60%%).\n<b>Pegout:</b> %s.\n<b>Bitcoin TX:</b> %s.\n<b>Runbook url:</b> %s", tonUrl2, btcUrl2, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (10 of 10 [100%]), next pegout",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress2,
						SigningMask:   new(big.Int).SetUint64(0b1111111111),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_2,
					}, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
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
			Name: "SEVERITY_OK, no pegouts",
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

	DoAlertTests(t, tests, NewAlertPegoutSigners())
}
