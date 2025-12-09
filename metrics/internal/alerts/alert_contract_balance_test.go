package alerts

import (
	"fmt"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertContractBalance(t *testing.T) {
	runbookUrl := mutils.CreateHTMLHyperlink("link", "http://runbook/ContractBalances.md")
	tonUrl := mutils.CreateHTMLHyperlink("link", "http://ton/0:0fb5045f7c7ad70131e8918305420ba9ba7c317d37b631a8a7d4560552dfbc78")
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
				Description: Description(fmt.Sprintf("The test_balance contract (%s) has a low balance: 1.000000000 TON.\n<b>Runbook url:</b> %s", tonUrl, runbookUrl)),
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
				Description: Description(fmt.Sprintf("The test_balance contract (%s) has a low balance: 0.333333333 TON.\n<b>Runbook url:</b> %s", tonUrl, runbookUrl)),
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertContractBalance("test", "test_balance", contractAddress))
}
