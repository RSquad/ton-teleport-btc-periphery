package alerts

import (
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

func TestAlertTotalServiceFee(t *testing.T) {
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (TotalServiceFee = 4000)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						TotalServiceFee: 4000,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (TotalServiceFee = 3000)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						TotalServiceFee: 3000,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (TotalServiceFee = 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						TotalServiceFee: 0,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (TotalServiceFee = -100)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						TotalServiceFee: -100,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: nil, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertTotalServiceFee())
}
