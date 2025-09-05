package alerts

import (
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func TestAlertDkgCulpritFound(t *testing.T) {
	labelsEmpty := Labels{
		"culprit_id": "",
		"is_evicted": "",
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK: No DKG",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labelsEmpty, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labelsEmpty, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labelsEmpty, Err: nil},
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
				DkgBeforeRestartDbFn: func(t time.Time) (*coordinator.DKG, error) {
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "1", "is_evicted": "YES"}, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL: 2 culprits, no reset",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "1", "is_evicted": "YES"}, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL: 2 culprits, reset",
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
				DkgBeforeRestartDbFn: func(t time.Time) (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "3,4", "is_evicted": "NO"}, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "3,4", "is_evicted": "NO"}, Err: nil},
		},
		{
			Name: "SEVERITY_OK: No DKG",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labelsEmpty, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgCulpritFound())
}
