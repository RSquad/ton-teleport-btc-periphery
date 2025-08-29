package alerts

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutRestarts(t *testing.T) {
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

	pegoutLabelsEmpty := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

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
				PegoutDbFn: func(address *address.Address) (*data_models.PegoutDbRow, error) {
					return &data_models.PegoutDbRow{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (no restart, just ExpiredAt != 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs, 0),
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
			Name: "SEVERITY_WARNING (1 restart)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+1, 0),
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
			Name: "SEVERITY_WARNING (2 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+2, 0),
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
			Name: "SEVERITY_WARNING (3 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+3, 0),
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
			Name: "SEVERITY_WARNING (4 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+4, 0),
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
			Name: "SEVERITY_WARNING (5 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+5, 0),
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
			Name: "SEVERITY_WARNING (6 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+6, 0),
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
			Name: "SEVERITY_WARNING (7 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+7, 0),
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
			Name: "SEVERITY_WARNING (8 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+8, 0),
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
			Name: "SEVERITY_WARNING (9 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+9, 0),
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
			Name: "SEVERITY_CRITICAL (10 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+10, 0),
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
			Name: "SEVERITY_OK (0 restarts, new pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress2,
						ExpiredAt:     time.Unix(0, 0),
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
			Name: "SEVERITY_OK (0 restarts, new pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutRestarts())
}
