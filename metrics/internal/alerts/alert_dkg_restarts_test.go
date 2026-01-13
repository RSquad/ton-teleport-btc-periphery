package alerts

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/mutils"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertDkgRestarts(t *testing.T) {
	validatorAddr1 := []byte{1, 2, 3, 4, 5}
	validatorAddr2 := []byte{6, 7, 8, 9, 10}
	validatorAddr3 := []byte{11, 12, 13, 14, 15}
	validatorAddr4 := []byte{16, 17, 18, 19, 20}
	validatorAddr5 := []byte{21, 22, 23, 24, 25}

	vSet := coordinator.VSet{
		0: validatorAddr1,
		1: validatorAddr2,
		2: validatorAddr3,
		3: validatorAddr4,
		4: validatorAddr5,
	}

	fullMask := new(big.Int).SetUint64(0b11111)

	maskWithout5 := new(big.Int).SetUint64(0b11110)

	maskWithout4And5 := new(big.Int).SetUint64(0b11100)

	maskOnlyFirstTwo := new(big.Int).SetUint64(0b11000)

	maskOnlyFirst := new(big.Int).SetUint64(0b10000)

	runbookUrl := mutils.CreateHTMLHyperlink("link", "http://runbook/DKGRestarts.md")

	testAddr, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")

	createTestDKG := func(vSetMask *big.Int) *coordinator.DKG {
		return &coordinator.DKG{
			VSet:       vSet,
			VSetMask:   vSetMask,
			MaxSigners: 5,
			State:      coordinator.DKGStateInProgress,
			Until:      time.Now().Add(time.Hour),
			Claims:     &coordinator.DKGClaims{},
			R1:         &coordinator.DKGR1{},
			R2:         &coordinator.DKGR2{},
			R3:         &coordinator.DKGR3{},
		}
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK - no restarts",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK - no DKG start events",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return nil, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_OK,
				Description: "OK",
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK - 1 timeout restart",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskWithout5),
						},
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
			Name: "SEVERITY_WARNING - 2 timeout restarts",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskWithout5),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 7},
								TxLT:    1002,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskWithout4And5),
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The DKG was restarted 2 times.\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL - 5 timeout restarts",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskWithout5),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 7},
								TxLT:    1002,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskWithout4And5),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 8},
								TxLT:    1003,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskOnlyFirstTwo),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 9},
								TxLT:    1004,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskOnlyFirst),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 10},
								TxLT:    1005,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskOnlyFirst),
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: Description(fmt.Sprintf("The DKG was restarted 5 times.\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_OK - 1 restart with evicted validator",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartValidatorEvicted,
							NewDkg: createTestDKG(maskWithout5),
						},
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
			Name: "SEVERITY_WARNING - 2 restarts with evicted validators",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartValidatorEvicted,
							NewDkg: createTestDKG(maskWithout5),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 7},
								TxLT:    1002,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartValidatorEvicted,
							NewDkg: createTestDKG(maskWithout4And5),
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The DKG was restarted 2 times.\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_WARNING - timeout expired with evicted validator",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartTimeoutExpired,
							NewDkg: createTestDKG(maskWithout5),
						},
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 7},
								TxLT:    1002,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartValidatorEvicted,
							NewDkg: createTestDKG(maskWithout4And5),
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_WARNING,
				Description: Description(fmt.Sprintf("The DKG was restarted 2 times.\n<b>Runbook url:</b> %s", runbookUrl)),
				Err:         nil,
			},
		},
		{
			Name: "SEVERITY_CRITICAL - unknown restart reason",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return []*coordinator.DKGRestartedEvent{
						{
							Raw: &ton.RawEvent{
								Addr:    testAddr,
								TxHash:  []byte{1, 2, 3, 4, 5, 6},
								TxLT:    1001,
								TxUtime: time.Now(),
							},
							Reason: coordinator.DkgRestartReason(999),
							NewDkg: createTestDKG(maskWithout5),
						},
					}, nil
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "",
				Err:         fmt.Errorf("unknown DKG restart reason 999"),
			},
		},
		{
			Name: "SEVERITY_CRITICAL - error while fetching last DKG start event",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return nil, fmt.Errorf("database error")
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "",
				Err:         fmt.Errorf("database error"),
			},
		},
		{
			Name: "SEVERITY_CRITICAL - error while fetching DKG restart events",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				EventsLastDkgStartedDBFn: func() (*coordinator.DKGStartedEvent, error) {
					return &coordinator.DKGStartedEvent{
						Raw: &ton.RawEvent{
							Addr:    testAddr,
							TxHash:  []byte{1, 2, 3, 4, 5},
							TxLT:    1000,
							TxUtime: time.Now(),
						},
						Dkg: createTestDKG(fullMask),
					}, nil
				},
				EventsAllFromDkgRestartDBFn: func(startLT uint64) ([]*coordinator.DKGRestartedEvent, error) {
					return nil, fmt.Errorf("restart events error")
				},
			}),
			Expect: TestResWant{
				Severity:    SEVERITY_CRITICAL,
				Description: "",
				Err:         fmt.Errorf("restart events error"),
			},
		},
	}

	DoAlertTests(t, tests, NewAlertDkgRestarts())
}
