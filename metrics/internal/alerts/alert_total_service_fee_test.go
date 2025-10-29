package alerts

import (
	"fmt"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

func TestAlertTotalServiceFee(t *testing.T) {
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/TotalServiceFee.md")
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
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (TotalServiceFee = 2999)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				TeleportContractStorageDbFn: func() (*teleportcontract.Storage, error) {
					return &teleportcontract.Storage{
						TotalServiceFee: 2999,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("Total service fee is less than 3000 satoshi.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Total service fee is less than 0 satoshi.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
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
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("Total service fee is less than 0 satoshi.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertTotalServiceFee())
}
