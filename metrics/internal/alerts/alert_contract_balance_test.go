package alerts

import (
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertContractBalance(t *testing.T) {
	contractAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	contractAddress2, _ := address.ParseAddr("Ef8VjV6LGTyiNLzefOm1dpuCMLcoewhqfQubtgbWcPwt2Gwp")

	labels := Labels{
		"address": contractAddress2.StringRaw(),
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK: Balance > 2",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalancesFn: func() (*data_models.ContractBalances, error) {
					return &data_models.ContractBalances{
						Balances: []*data_models.ContractBalance{
							{
								Name:    "test2",
								Addr:    contractAddress1,
								Balance: 1 * 1000000000,
							},
							{
								Name:    "test1",
								Addr:    contractAddress2,
								Balance: 3 * 1000000000,
							},
						},
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING: 0.5 < Balance < 2",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalancesFn: func() (*data_models.ContractBalances, error) {
					return &data_models.ContractBalances{
						Balances: []*data_models.ContractBalance{
							{
								Name:    "test2",
								Addr:    contractAddress1,
								Balance: 1 * 1000000000,
							},
							{
								Name:    "test1",
								Addr:    contractAddress2,
								Balance: 1 * 1000000000,
							},
						},
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL: Balance < 0.5",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				ActualContractBalancesFn: func() (*data_models.ContractBalances, error) {
					return &data_models.ContractBalances{
						Balances: []*data_models.ContractBalance{
							{
								Name:    "test2",
								Addr:    contractAddress1,
								Balance: 1 * 1000000000,
							},
							{
								Name:    "test1",
								Addr:    contractAddress2,
								Balance: 1000000000 / 3,
							},
						},
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: labels, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertContractBalance("test1"))
}
