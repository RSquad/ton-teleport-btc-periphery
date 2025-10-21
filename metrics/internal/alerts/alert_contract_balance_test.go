package alerts

import (
	"testing"

	"github.com/xssnick/tonutils-go/address"
)

func TestAlertContractBalance(t *testing.T) {
	contractAddress, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK: Balance > 2",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalanceFn: func(name string) (int64, error) {
					return 3 * 1000000000, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING: 0.5 < Balance < 2",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalanceFn: func(name string) (int64, error) {
					return 1 * 1000000000, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: "The test_balance contract (<a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>) has a low balance: 1.000000000 TON.",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL: Balance < 0.5",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalanceFn: func(name string) (int64, error) {
					return 1000000000 / 3, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "The test_balance contract (<a href=\"http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78\">0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78</a>) has a low balance: 0.333333333 TON.",
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertContractBalance("test", "test_balance", contractAddress))
}
