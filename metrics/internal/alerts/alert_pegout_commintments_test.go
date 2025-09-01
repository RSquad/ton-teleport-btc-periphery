package alerts

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutCommintments(t *testing.T) {
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

	pegoutLabelsEmpty := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_INFO, Labels: pegoutLabels1, Err: nil},
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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: pegoutLabels1, Err: nil},
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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
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
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_2,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels2, Err: nil},
		},
		{
			Name: "SEVERITY_OK, no pegouts",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return nil, nil
				},
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutCommintments())
}
