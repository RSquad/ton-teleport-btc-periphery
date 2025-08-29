package alerts

import (
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/xssnick/tonutils-go/address"
)

func TestAlertPegoutSigningDuration(t *testing.T) {
	signingTimeout := int64(60 * 20) // 20 min
	beginTs := int64(123456)
	pegoutAddress1, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	pegoutAddress2, _ := address.ParseAddr("Ef8VjV6LGTyiNLzefOm1dpuCMLcoewhqfQubtgbWcPwt2Gwp")

	pegoutLabels1 := Labels{
		"bitcoin_tx_id": "unknonwn",
		"pegout_addr":   pegoutAddress1.StringRaw(),
	}

	pegoutLabels2 := Labels{
		"bitcoin_tx_id": "unknonwn",
		"pegout_addr":   pegoutAddress2.StringRaw(),
	}

	pegoutLabelsEmpty := Labels{
		"bitcoin_tx_id": "",
		"pegout_addr":   "",
	}

	tests := []TestDesc{
		{
			Name: "SEVERITY_OK (new unsigned pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(0, 0),
					}, nil
				},
				CoordinatorContractDataDbFn: func() (*coordinator.Storage, error) {
					return &coordinator.Storage{
						SigningTimeout: uint32(signingTimeout),
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (same unsigned pegout, 1 minute later)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*1
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_OK (same unsigned pegout, 11 minute later)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*11
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (same unsigned pegout, 20 minute later)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*20
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (same unsigned pegout, 2 minute later after restart, 22 minutes total)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress1,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*2, 0),
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*22
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels1, Err: nil},
		},
		{
			Name: "SEVERITY_CRITICAL (new unsigned pegout, Its still SEVERITY_CRITICAL, and we dont know how much time it will take to sign the current pegout)",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					beginTs = beginTs + 60*24 // Update beginTs for new pegout pegoutAddress2
					return &coordinator.PegoutRecord{
						PegoutAddress: pegoutAddress2,
						ExpiredAt:     time.Unix(beginTs+signingTimeout*1, 0),
					}, nil
				},
				NowUnixTsFn: func() int64 {
					return beginTs + 60*25
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_CRITICAL, Labels: pegoutLabels2, Err: nil},
		},
		{
			Name: "SEVERITY_OK, all pegout are signed",
			DataSource: NewAlertDataSourceTesting(AlertDataSourceTestingConfig{
				FirstUnsignedPegoutDbFn: func() (*coordinator.PegoutRecord, error) {
					return nil, nil
				},
			}),
			Expect: TestResWant{Severity: SEVERITY_OK, Labels: pegoutLabelsEmpty, Err: nil},
		},
	}

	DoAlertTests(t, tests, NewAlertPegoutSigningDuration())
}
