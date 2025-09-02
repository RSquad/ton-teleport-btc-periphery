package alerts

import (
	"math/big"
	"testing"

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
			Name: "SEVERITY_OK: DKG, no culprit",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
						State:      coordinator.DKGStateInProgress,
						Claims: &coordinator.DKGClaims{
							Counters: make(coordinator.DKGClaimcounters, 0),
							Mask:     big.NewInt(0b0000000000),
						},
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						MinClaimsPercent: 66,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labelsEmpty, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL: DKG, 1 culprit, evicted",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
					counters[1] = 7

					return &coordinator.DKG{
						MaxSigners: 10,
						State:      coordinator.DKGStateInProgress,
						Claims: &coordinator.DKGClaims{
							Counters: counters,
							Mask:     big.NewInt(0b0000000010),
						},
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						MinClaimsPercent: 66,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "1", "is_evicted": "YES"}, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL: DKG, 2 culprit, evicted no",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)
					counters[1] = 7
					counters[3] = 5

					return &coordinator.DKG{
						MaxSigners: 10,
						State:      coordinator.DKGStateInProgress,
						Claims: &coordinator.DKGClaims{
							Counters: counters,
							Mask:     big.NewInt(0b0000001010),
						},
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						MinClaimsPercent: 66,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: Labels{"culprit_id": "3", "is_evicted": "NO"}, Err: nil},
		},
		{
			Name: "SEVERITY_OK: DKG, 0 culprit, next DKG round",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					counters := make(coordinator.DKGClaimcounters)

					return &coordinator.DKG{
						MaxSigners: 10,
						State:      coordinator.DKGStatePart1Finished,
						Claims: &coordinator.DKGClaims{
							Counters: counters,
							Mask:     big.NewInt(0b0000000000),
						},
					}, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						MinClaimsPercent: 66,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labelsEmpty, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgCulpritFound())
}
