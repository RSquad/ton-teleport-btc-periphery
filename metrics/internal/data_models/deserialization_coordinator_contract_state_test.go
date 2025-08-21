package data_models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

func mustUnmarshalJSONMap(t *testing.T, js string) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(js), &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return m
}

func TestDeserializeCoordinatorContractState_Success(t *testing.T) {
	js := `{
		"Initiated": true,
		"StandaloneMode": true,
		"Id": 266458210,
		"ConfiguratorAddr": "0:0f06dd725549dd60a6ca3743ac532db52789cb15af1f943fe96adbb4c1a86155",
		"Enabled": true,
		"UnsignedPegouts": [{
			"ID": 135,
			"PegoutAddress": "EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-",
			"InternalKey": "ZwqZwANOMRDlrCsHEsOeem9vDFljXk5IDh4H+OMhDTk=",
			"IsAutopegout": true,
			"Commitments": {
				"46": "AQAjD4qzA+NRZBmoOjp0ftul6qYqcwR147bphPYltmE5VLIYCxcOAz18T35KTKHp0+O5RHk6kI+ni5qHIw3Mkf6PZgZ1wFoB",
				"90": "AQAjD4qzA90LtHszx0UfOLPxR1Z9LvFVH5MkVPb/5sig+UelyJdrAh6h19hJkHvRee0qpYlx/rH6Cx0i0VHvq6dUMvq76NSf"
			},
			"CommitmentsMaskAccepted": "0",
			"CommitmentsMaskOther": "1237940039285450643643301888",
			"SigningShares": {},
			"SigningSharesMask": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			"Signatures": {
				"Mask": "0",
				"Count": 0,
				"Hash": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
			},
			"ClaimsMask": "0",
			"ClaimsCount": 0,
			"ClaimsCounters": {},
			"MaxSigners": 91,
			"ExpiredAt": "2025-07-16T03:27:21Z",
			"SigningMask": "2475880078570760549798248447"
		}],
		"MinClaimsPercent": 51,
		"MinSignersThreshold": 51,
		"DkgLifetime": 300,
		"SigningTimeout": 300,
		"NextPegoutIdx": 135,
		"TeleportAddr": "0:6d69b0ff0cb6f1883aa2829b174dc406b88a7ff08dba9a1b31399058a2a41a70"
	}`

	m := mustUnmarshalJSONMap(t, js)

	got, err := DeserializeCoordinatorContractState(m)
	if err != nil {
		t.Fatalf("DeserializeCoordinatorContractState error: %v", err)
	}

	// Simple scalars
	if !got.Initiated {
		t.Errorf("Initiated=false, want true")
	}
	if !got.StandaloneMode {
		t.Errorf("StandaloneMode=false, want true")
	}
	if got.Id != 266458210 {
		t.Errorf("Id=%d, want 266458210", got.Id)
	}
	if !got.Enabled {
		t.Errorf("Enabled=false, want true")
	}
	if got.MinClaimsPercent != 51 {
		t.Errorf("MinClaimsPercent=%d, want 51", got.MinClaimsPercent)
	}
	if got.MinSignersThreshold != 51 {
		t.Errorf("MinSignersThreshold=%d, want 51", got.MinSignersThreshold)
	}
	if got.DkgLifetime != 300 {
		t.Errorf("DkgLifetime=%d, want 300", got.DkgLifetime)
	}
	if got.SigningTimeout != 300 {
		t.Errorf("SigningTimeout=%d, want 300", got.SigningTimeout)
	}
	if got.NextPegoutIdx != 135 {
		t.Errorf("NextPegoutIdx=%d, want 135", got.NextPegoutIdx)
	}

	// Addresses
	wantCfg, _ := address.ParseRawAddr("0:0f06dd725549dd60a6ca3743ac532db52789cb15af1f943fe96adbb4c1a86155")
	if got.ConfiguratorAddr == nil || got.ConfiguratorAddr.String() != wantCfg.String() {
		t.Errorf("ConfiguratorAddr=%v, want %v", got.ConfiguratorAddr, wantCfg)
	}
	wantTele, _ := address.ParseRawAddr("0:6d69b0ff0cb6f1883aa2829b174dc406b88a7ff08dba9a1b31399058a2a41a70")
	if got.TeleportAddr == nil || got.TeleportAddr.String() != wantTele.String() {
		t.Errorf("TeleportAddr=%v, want %v", got.TeleportAddr, wantTele)
	}

	// UnsignedPegouts is wired through and parsed (deep details already covered by its own tests)
	if len(got.UnsignedPegouts) != 1 {
		t.Fatalf("UnsignedPegouts len=%d, want 1", len(got.UnsignedPegouts))
	}
	if got.UnsignedPegouts[0].ID != 135 {
		t.Errorf("UnsignedPegouts[0].ID=%d, want 135", got.UnsignedPegouts[0].ID)
	}

	// Quick sanity for a nested timestamp (parsing exercised in DeserializePegouts)
	wantTime := time.Date(2025, 7, 16, 3, 27, 21, 0, time.UTC)
	if !got.UnsignedPegouts[0].ExpiredAt.Equal(wantTime) {
		t.Errorf("UnsignedPegouts[0].ExpiredAt=%v, want %v", got.UnsignedPegouts[0].ExpiredAt, wantTime)
	}
}

func TestDeserializeCoordinatorContractState_EmptyOK(t *testing.T) {
	m := map[string]interface{}{}

	got, err := DeserializeCoordinatorContractState(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All zero-values / nil pointers expected
	if got.Initiated || got.StandaloneMode || got.Enabled {
		t.Errorf("boolean fields should be false by default: %+v", got)
	}
	if got.Id != 0 || got.MinClaimsPercent != 0 || got.MinSignersThreshold != 0 ||
		got.DkgLifetime != 0 || got.SigningTimeout != 0 || got.NextPegoutIdx != 0 {
		t.Errorf("numeric fields not zero by default: %+v", got)
	}
	if got.ConfiguratorAddr != nil || got.TeleportAddr != nil {
		t.Errorf("address fields should be nil by default: %+v", got)
	}
	if len(got.UnsignedPegouts) != 0 {
		t.Errorf("UnsignedPegouts len=%d, want 0", len(got.UnsignedPegouts))
	}
}

func TestDeserializeCoordinatorContractState_Errors(t *testing.T) {
	t.Run("Initiated bad type", func(t *testing.T) {
		m := map[string]interface{}{"Initiated": "nope"}
		_, err := DeserializeCoordinatorContractState(m)
		if err == nil || !strings.Contains(err.Error(), "`Initiated` parse error") {
			t.Fatalf("want Initiated parse error, got: %v", err)
		}
	})

	t.Run("Id not a number", func(t *testing.T) {
		m := map[string]interface{}{"Id": "abc"}
		_, err := DeserializeCoordinatorContractState(m)
		if err == nil || !strings.Contains(err.Error(), "`Id` parse error") {
			t.Fatalf("want Id parse error, got: %v", err)
		}
	})

	t.Run("ConfiguratorAddr bad format", func(t *testing.T) {
		m := map[string]interface{}{"ConfiguratorAddr": "not-an-addr"}
		_, err := DeserializeCoordinatorContractState(m)
		if err == nil || !strings.Contains(err.Error(), "`ConfiguratorAddr` parse error") {
			t.Fatalf("want ConfiguratorAddr parse error, got: %v", err)
		}
	})

	t.Run("TeleportAddr bad format", func(t *testing.T) {
		m := map[string]interface{}{"TeleportAddr": "bad-addr"}
		_, err := DeserializeCoordinatorContractState(m)
		if err == nil || !strings.Contains(err.Error(), "`TeleportAddr` parse error") {
			t.Fatalf("want TeleportAddr parse error, got: %v", err)
		}
	})

	t.Run("UnsignedPegouts propagates nested error", func(t *testing.T) {
		// element missing required ID should error in DeserializePegouts
		js := `{"UnsignedPegouts":[{"PegoutAddress":"EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-"}]}`
		m := mustUnmarshalJSONMap(t, js)
		_, err := DeserializeCoordinatorContractState(m)
		if err == nil || !strings.Contains(err.Error(), "`UnsignedPegouts` parse error") {
			t.Fatalf("want UnsignedPegouts parse error, got: %v", err)
		}
	})

	t.Run("NextPegoutIdx wrong type", func(t *testing.T) {
		m := map[string]interface{}{"NextPegoutIdx": "oops"}
		_, err := DeserializeCoordinatorContractState(m)
		if err == nil || !strings.Contains(err.Error(), "`NextPegoutIdx` parse error") {
			t.Fatalf("want NextPegoutIdx parse error, got: %v", err)
		}
	})
}
