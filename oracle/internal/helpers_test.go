package helpers

import (
	"testing"
)

func TestCalcMinSigners(t *testing.T) {
	tests := []struct {
		name       string
		maxSigners uint16
		want       uint16
		wantErr    bool
	}{
		{
			name:       "maxSigners = 0",
			maxSigners: 0,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "maxSigners = 1",
			maxSigners: 1,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "maxSigners = 2",
			maxSigners: 2,
			want:       2,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 3",
			maxSigners: 3,
			want:       2,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 4",
			maxSigners: 4,
			want:       2,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 5",
			maxSigners: 5,
			want:       3,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 6",
			maxSigners: 6,
			want:       4,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 7",
			maxSigners: 7,
			want:       4,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 9",
			maxSigners: 9,
			want:       6,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcMinSigners(tt.maxSigners)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalcMinSigners(%d) error = %v, wantErr %v", tt.maxSigners, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("CalcMinSigners(%d) = %d, want %d", tt.maxSigners, got, tt.want)
			}
		})
	}
}

func TestExtractExitCode(t *testing.T) {
	tests := []struct {
		name         string
		errorLog     string
		expectedCode int
		expectError  bool
	}{
		{
			name:         "Extract exitcode from standard format",
			errorLog:     "Cannot run message on account: inbound external message rejected by transaction B0279A0FE4EF5A2759AFDCD421FF5213215357E9D932564B9B921B9B88EC6018: exitcode=114, steps=133, gas_used=0",
			expectedCode: 114,
			expectError:  false,
		},
		{
			name:         "No exitcode in error log",
			errorLog:     "Some other error without exitcode",
			expectedCode: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := ExtractExitCode(tt.errorLog)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if code != tt.expectedCode {
					t.Errorf("Expected exitcode %d, got %d", tt.expectedCode, code)
				}
			}
		})
	}
}
