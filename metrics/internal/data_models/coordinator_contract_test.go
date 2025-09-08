package data_models

import (
	"testing"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

func TestDeserializeCoordinatorContractStorage_Success(t *testing.T) {
	jsonInput := `{
		"Initiated": true,
		"StandaloneMode": true,
		"Id": 266458210,
		"ConfiguratorAddr": "EQAPBt1yVUndYKbKN0OsUy21J4nLFa8flD_patu0wahhVQmA",
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
			"CommitmentsMaskAccepted": 0,
			"CommitmentsMaskOther": 1237940039285450643643301888,
			"SigningShares": {},
			"SigningSharesMask": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			"Signatures": {
				"Mask": 0,
				"Count": 0,
				"Hash": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
			},
			"ClaimsMask": 0,
			"ClaimsCount": 0,
			"ClaimsCounters": {},
			"MaxSigners": 91,
			"ExpiredAt": "2025-07-16T03:27:21Z",
			"SigningMask": 2475880078570760549798248447
		}],
		"MinClaimsPercent": 51,
		"MinSignersThreshold": 51,
		"DkgLifetime": 300,
		"SigningTimeout": 300,
		"NextPegoutIdx": 135,
		"TeleportAddr": "EQBtabD_DLbxiDqigpsXTcQGuIp_8I26mhsxOZBYoqQacImn"
	}`

	got, err := DeserializeCoordinatorContractStorage([]byte(jsonInput))
	if err != nil {
		t.Fatalf("DeserializeCoordinatorContractStorage error: %v", err)
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
	wantCfg, _ := address.ParseAddr("EQAPBt1yVUndYKbKN0OsUy21J4nLFa8flD_patu0wahhVQmA")
	if got.ConfiguratorAddr == nil || got.ConfiguratorAddr.String() != wantCfg.String() {
		t.Errorf("ConfiguratorAddr=%v, want %v", got.ConfiguratorAddr, wantCfg)
	}
	wantTele, _ := address.ParseAddr("EQBtabD_DLbxiDqigpsXTcQGuIp_8I26mhsxOZBYoqQacImn")
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
