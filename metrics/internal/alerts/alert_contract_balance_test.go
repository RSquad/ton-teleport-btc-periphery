package alerts

import (
	"testing"

	"github.com/xssnick/tonutils-go/address"
)

func TestAlertContractBalance(t *testing.T) {
	contractAddress, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")

	labels := Labels{
		"address": contractAddress.StringRaw(),
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK: Balance > 2",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalanceFn: func(name string) (int64, error) {
					return 3 * 1000000000, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING: 0.5 < Balance < 2",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalanceFn: func(name string) (int64, error) {
					return 1 * 1000000000, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL: Balance < 0.5",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalanceFn: func(name string) (int64, error) {
					return 1000000000 / 3, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: labels, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertContractBalance("test", "test_balance", contractAddress))
}
