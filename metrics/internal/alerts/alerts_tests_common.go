package alerts

import (
	"testing"

	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

type TestResWant struct {
	Severity Severity
	Labels   []string
	Err      error
}

type TestDesc struct {
	Name       string
	DataSource AlertDataSource
	Expect     TestResWant
}

func DoAlertTests(t *testing.T, tests []TestDesc, alert Alert) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			severity, labels, err := alert.Check(tt.DataSource)

			// Assert
			if tt.Expect.Err != nil {
				if err == nil || err.Error() != tt.Expect.Err.Error() {
					t.Fatalf("expected error %v, got %v", tt.Expect.Err, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if severity != tt.Expect.Severity {
				t.Fatalf("expected severity %v, got %v", tt.Expect.Severity, severity)
			}

			if !mutils.IsEqual(labels, tt.Expect.Labels) {
				t.Fatalf("expected labels %v, got %v", tt.Expect.Labels, labels)
			}
		})
	}
}
