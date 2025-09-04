package alerts

import (
	"context"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func TestAlertDkgParticipants(t *testing.T) {
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (DKG with 100% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 100,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_OK (DKG with 90% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 90,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (DKG with 80% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 80,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (DKG with 70% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 70,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (DKG with 60% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 60,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (DKG with 55% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 55,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: nil, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (DKG with 51% participants)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				PrevDkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 51,
					}, nil
				},
				TonMaxMainValidatorsFn: func(ctx context.Context) (int, error) {
					return 100, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: nil, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgParticipants())
}
