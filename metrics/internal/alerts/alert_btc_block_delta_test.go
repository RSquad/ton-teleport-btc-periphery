package alerts

import (
	"testing"
)

func TestAlertBtcBlockDelta(t *testing.T) {
	labels := Labels{}
	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (delta == 0)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 0, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_WARNING (delta == 1)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 1, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_WARNING, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (delta == 2)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 2, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (delta > 2)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				BtcGetBestBlockHeightFn: func() (int, error) {
					return 3, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: labels, Err: nil},
		},
		// TODO: add shard chain tests
	}

	DoAlertTests(t, tests, NewAlertPegoutInMempool())
}
