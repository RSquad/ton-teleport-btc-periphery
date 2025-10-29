package alerts

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

func TestAlertDkgCulpritFound(t *testing.T) {
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/DKGCulprit.md")
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK: No DKG",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK: First DKG try",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until:    time.Unix(1, 0),
						VSetMask: big.NewInt(0b1111111111),
						Claims: &coordinator.DKGClaims{
							Counters: make(coordinator.DKGClaimcounters, 0),
						},
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
			Name: "SEVERITY_OK: 1 culprit, no reset",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
					counters[1] = 7

					return &coordinator.DKG{
						Until:    time.Unix(1, 0),
						VSetMask: big.NewInt(0b1111111111),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
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
			Name: "SEVERITY_CRITICAL: 1 culprit, reset",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)

					return &coordinator.DKG{
						Until:    time.Unix(2, 0),
						VSetMask: big.NewInt(0b1111111101),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
					}, nil
				},
				DkgUntilDbFn: func(t time.Time) (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
					counters[1] = 7

					return &coordinator.DKG{
						Until:    time.Unix(1, 0),
						VSetMask: big.NewInt(0b1111111111),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						MinClaimsPercent: 51,
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
				Description: Description(fmt.Sprintf("DKG culprit found. Culprit ids: [1]. Not evicted ids: []. Runbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL: 1 real and 2 potential culprits, no reset",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
					counters[1] = 7
					counters[3] = 4
					counters[4] = 4

					return &coordinator.DKG{
						Until:    time.Unix(2, 0),
						VSetMask: big.NewInt(0b1111111101),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("DKG culprit found. Culprit ids: [1]. Not evicted ids: []. Runbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL: 2 real and 2 potential culprits, reset",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)

					return &coordinator.DKG{
						Until:    time.Unix(3, 0),
						VSetMask: big.NewInt(0b1111111101),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
					}, nil
				},
				DkgUntilDbFn: func(t time.Time) (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
					counters[1] = 7
					counters[3] = 4
					counters[4] = 4

					return &coordinator.DKG{
						Until:    time.Unix(2, 0),
						VSetMask: big.NewInt(0b1111111101),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						MinClaimsPercent: 51,
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
				Description: Description(fmt.Sprintf("DKG culprit found. Culprit ids: [1]. Not evicted ids: [3,4]. Runbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL: 0 culprits, no reset",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)

					return &coordinator.DKG{
						Until:    time.Unix(3, 0),
						VSetMask: big.NewInt(0b1111111101),
						Claims: &coordinator.DKGClaims{
							Counters: counters,
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("DKG culprit found. Culprit ids: [1]. Not evicted ids: [3,4]. Runbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK: No DKG",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
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

	DoAlertTests(t, tests, NewAlertDkgCulpritFound())
}
