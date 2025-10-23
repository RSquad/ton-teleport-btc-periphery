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

	bitcoin_tx_id_1, _ := hex.DecodeString("f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b")
	bitcoin_tx_id_2, _ := hex.DecodeString("3d46303861d5336c3ebdea3a20883a1cb77f4f3a66a2fb5e6494d3a0ab878bd1")

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
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
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
			Name: "SEVERITY_OK (no restart, just ExpiredAt != 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
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
			Name: "SEVERITY_WARNING (1 restart)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+1, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: "The pegout signing was restarted 1 times. Pegout: <a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>. Bitcoin TX: <a href=\"http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b\">f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b</a>",
				Err:         nil,
			},
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
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: "The pegout signing was restarted 2 times. Pegout: <a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>. Bitcoin TX: <a href=\"http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b\">f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b</a>",
				Err:         nil,
			},
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
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: "The pegout signing was restarted 3 times. Pegout: <a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>. Bitcoin TX: <a href=\"http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b\">f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b</a>",
				Err:         nil,
			},
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
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: "The pegout signing was restarted 4 times. Pegout: <a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>. Bitcoin TX: <a href=\"http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b\">f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b</a>",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (5 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+5, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "The pegout signing was restarted 5 times. Pegout: <a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>. Bitcoin TX: <a href=\"http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b\">f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b</a>",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (6 restarts)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+6, 0),
					}, nil
				},
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_1,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "The pegout signing was restarted 6 times. Pegout: <a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>. Bitcoin TX: <a href=\"http://btc/f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b\">f7df2a86684e500a3c6c7ca785b8500e4e3c89d1751edf86b6deb68e761a329b</a>",
				Err:         nil,
			},
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
				PegoutDbFn: func(address *address.Address) (*data_models.Pegout, error) {
					return &data_models.Pegout{
						BitcoinTxId: bitcoin_tx_id_2,
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
			Name: "SEVERITY_OK (0 restarts, new pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
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

	DoAlertTests(t, tests, NewAlertPegoutRestarts())
}
