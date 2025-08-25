package alerts

import (
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func TestAlertPegoutMaxSignersCount(t *testing.T) {
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (10 of 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0111101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0111011100111),
					}, nil
				},
				PrevDkgFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: []string{}, Err: nil},
		},
		{
			Name: "SEVERITY_INFO (9 of 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0111101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0111011100101),
					}, nil
				},
				PrevDkgFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_INFO, Labels: []string{}, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (8 of 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0101101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0101011100101),
					}, nil
				},
				PrevDkgFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: []string{}, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (7 of 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0001101100101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0001011100101),
					}, nil
				},
				PrevDkgFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: []string{}, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (6 of 10)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						CommitmentsMaskAccepted: new(big.Int).SetUint64(0b0001101000101),
						CommitmentsMaskOther:    new(big.Int).SetUint64(0b0001011000101),
					}, nil
				},
				PrevDkgFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						MaxSigners: 10,
					}, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: []string{}, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutMaxSignersCount())
}
