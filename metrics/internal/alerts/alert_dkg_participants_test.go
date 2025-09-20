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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: nil, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: nil, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: nil, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: nil, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: nil, Err: nil},
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
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: nil, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgParticipants())
}
