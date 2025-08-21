package alerts

import (
	"math/big"
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

type want struct {
	severity Severity
	err      error
}

func TestAlertPegoutMaxSignersCount_Check(t *testing.T) {
	tests := []struct {
		name       string
		dataSource AlertDataSource
		expect     want
	}{
		{
			name: "SEVERITY_OK (10 of 10)",
			dataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
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
			expect: want{severity: SEVERITY_OK, err: nil},
		},
		{
			name: "SEVERITY_INFO (9 of 10)",
			dataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
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
			expect: want{severity: SEVERITY_INFO, err: nil},
		},
		{
			name: "SEVERITY_WARNING (8 of 10)",
			dataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
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
			expect: want{severity: SEVERITY_WARNING, err: nil},
		},
		{
			name: "SEVERITY_CRITICAL (7 of 10)",
			dataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
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
			expect: want{severity: SEVERITY_CRITICAL, err: nil},
		},
		{
			name: "SEVERITY_CRITICAL (6 of 10)",
			dataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
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
			expect: want{severity: SEVERITY_CRITICAL, err: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := NewAlertPegoutMaxSignersCount()
			severity, err := alert.Check(tt.dataSource)

			// Assert
			if tt.expect.err != nil {
				if err == nil || err.Error() != tt.expect.err.Error() {
					t.Fatalf("expected error %v, got %v", tt.expect.err, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if severity != tt.expect.severity {
				t.Fatalf("expected severity %v, got %v", tt.expect.severity, severity)
			}
		})
	}
}
