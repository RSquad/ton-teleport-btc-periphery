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

func TestAlertPegoutCommintments(t *testing.T) {
	pegoutAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	pegoutAddress2, _ := address.ParseAddr("Ef8VjV6LGTyiNLzefOm1dpuCMLcoewhqfQubtgbWcPwt2Gwp")

	bitcoin_tx_id_1, _ := hex.DecodeString("f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	bitcoin_tx_id_2, _ := hex.DecodeString("3d46303861d5336c3ebdea3a20883a1cb77f4f3a66a2fb5e6494d3a0ab878bd1")

	tonUrl := mutils.CreateShortLink("link", "http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78")
	btcUrl := mutils.CreateShortLink("link", "http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/PegoutCommitments.md")

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (4 of 10 [40%], Commitment stage)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress:           pegoutAddress1,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0111100000000),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b01110100000000),
						Signatures:              coordinator.PegoutSignatures{Count: 0},
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
			Name: "SEVERITY_OK (10 of 10 [100%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress:           pegoutAddress1,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0111101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0111011100111),
						Signatures:              coordinator.PegoutSignatures{Count: 1},
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
						PegoutAddress:           pegoutAddress1,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0111101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0111011100101),
						Signatures:              coordinator.PegoutSignatures{Count: 1},
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
				Description: Description(fmt.Sprintf("The number of pegout commitments is 9 of 10 (90%%). Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (8 of 10 [80%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress:           pegoutAddress1,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0101101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0101011100101),
						Signatures:              coordinator.PegoutSignatures{Count: 1},
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
				Description: Description(fmt.Sprintf("The number of pegout commitments is 8 of 10 (80%%). Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (7 of 10 [70%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress:           pegoutAddress1,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0001101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0001011100101),
						Signatures:              coordinator.PegoutSignatures{Count: 1},
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
				Description: Description(fmt.Sprintf("The number of pegout commitments is 7 of 10 (70%%). Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (6 of 10 [60%])",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress:           pegoutAddress1,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0001101000101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0001011000101),
						Signatures:              coordinator.PegoutSignatures{Count: 1},
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
				Description: Description(fmt.Sprintf("The number of pegout commitments is 6 of 10 (60%%). Pegout: %s. Bitcoin TX: %s. Runbook url: %s", tonUrl, btcUrl, runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (4 of 10 [40%], Commitment stage, new pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress:           pegoutAddress2,
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0111100000000),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b01110100000000),
						Signatures:              coordinator.PegoutSignatures{Count: 0},
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

	DoAlertTests(t, tests, NewAlertPegoutCommintments())
}
