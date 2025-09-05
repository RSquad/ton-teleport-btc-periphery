package alerts

import (
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func TestAlertDkgCulpritFound(t *testing.T) {
	labelsEmpty := Labels{
		"culprit_id":      "",
		"not_evicted_ids": "",
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "1", "not_evicted_ids": ""}, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "1", "not_evicted_ids": ""}, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "", "not_evicted_ids": "3,4"}, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "", "not_evicted_ids": "3,4"}, Err: nil},
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
