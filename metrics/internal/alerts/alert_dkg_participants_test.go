package alerts

import (
	"context"
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func TestAlertDkgParticipants(t *testing.T) {
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (DKG with 100% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					vset := make(coordinator.VSet, 10)
					for i := range 10 {
						vset[uint16(i)] = nil
					}

					return &coordinator.DKG{
						State:    coordinator.DKGStateFinished,
						VSet:     vset,
						VSetMask: big.NewInt(0b1111111111),
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 10, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						StandaloneMode: false,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "The number of DKG participants is 10 of 10 (100%)",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK(DKG with 90% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					vset := make(coordinator.VSet, 10)
					for i := range 10 {
						vset[uint16(i)] = nil
					}

					return &coordinator.DKG{
						State:    coordinator.DKGStateFinished,
						VSet:     vset,
						VSetMask: big.NewInt(0b0111111111),
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 10, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						StandaloneMode: false,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "The number of DKG participants is 9 of 10 (90%)",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (DKG with 80% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					vset := make(coordinator.VSet, 10)
					for i := range 10 {
						vset[uint16(i)] = nil
					}

					return &coordinator.DKG{
						State:    coordinator.DKGStateFinished,
						VSet:     vset,
						VSetMask: big.NewInt(0b0111111101),
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 10, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						StandaloneMode: false,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: "The number of DKG participants is 8 of 10 (80%). Steps to resolve: https://github.com/RSquad/teleport-runbooks/blob/master/alerts/DKGParticipants.md",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (DKG with 50% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					vset := make(coordinator.VSet, 10)
					for i := range 10 {
						vset[uint16(i)] = nil
					}

					return &coordinator.DKG{
						State:    coordinator.DKGStateFinished,
						VSet:     vset,
						VSetMask: big.NewInt(0b0101010101),
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 10, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						StandaloneMode: false,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "The number of DKG participants is 5 of 10 (50%). Steps to resolve: https://github.com/RSquad/teleport-runbooks/blob/master/alerts/DKGParticipants.md",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL from old DKG (new DKG in progress)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					vset := make(coordinator.VSet, 10)
					for i := range 10 {
						vset[uint16(i)] = nil
					}

					return &coordinator.DKG{
						State:    coordinator.DKGStateInProgress,
						VSet:     vset,
						VSetMask: big.NewInt(0b01111111111),
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 10, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						StandaloneMode: false,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "The number of DKG participants is 5 of 10 (50%). Steps to resolve: https://github.com/RSquad/teleport-runbooks/blob/master/alerts/DKGParticipants.md",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (DKG with 100% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					vset := make(coordinator.VSet, 10)
					for i := range 10 {
						vset[uint16(i)] = nil
					}

					return &coordinator.DKG{
						State:    coordinator.DKGStateFinished,
						VSet:     vset,
						VSetMask: big.NewInt(0b1111111111),
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 10, nil
				},
				CoordinatorContractStorageDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						StandaloneMode: false,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "The number of DKG participants is 10 of 10 (100%)",
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgParticipants())
}
