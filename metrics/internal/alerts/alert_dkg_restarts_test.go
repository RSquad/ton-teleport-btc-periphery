package alerts

import (
	"fmt"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
)

func TestAlertDkgRestarts(t *testing.T) {
	runbookUrl := mutils.CreateShortLink("link", "http://runbook/DKGRestarts.md")
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (new DKG)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(1, 0),
						State: coordinator.DKGStateInProgress,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (restarts: 1)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(2, 0),
						State: coordinator.DKGStateInProgress,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (restarts: 2)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(3, 0),
						State: coordinator.DKGStatePart2Finished,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The DKG was restarted 2 times.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (restarts: 3)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(4, 0),
						State: coordinator.DKGStatePart2Finished,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The DKG was restarted 3 times.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING (restarts: 4)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(5, 0),
						State: coordinator.DKGStatePart2Finished,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The DKG was restarted 4 times.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (restarts: 5)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(6, 0),
						State: coordinator.DKGStateInProgress,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("The DKG was restarted 5 times.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL (restarts: 6)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(7, 0),
						State: coordinator.DKGStatePart2Finished,
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("The DKG was restarted 6 times.\nRunbook url: %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (no DKG)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK (new DKG)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				DkgDbFn: func() (*coordinator.DKG, error) {
					return &coordinator.DKG{
						Until: time.Unix(100, 0),
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgRestarts())
}
