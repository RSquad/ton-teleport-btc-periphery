package data_models

import (
	"bytes"
	"encoding/base64"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func TestDeserializeDkg_Success(t *testing.T) {
	js := `{
		"State": 0,
		"VSet": {
			"0": "x4UOi03GzPdmCyhlvnT9MGINIXctTdR4brnf6yIReuQ=",
			"1": "bODZpBjiY/TkPsHKhmGIBR7hTkS5TkGh8qQR1IBKsGs=",
			"10": "1V7vbGGD7VbJtVV6pnChySgh3Y0SRHuQXI9e9h/XhxU=",
			"11": "9XGt1LzAEORqoVUFgUvZGnjyzYuqymfRMu8G3nzhTLw="
		},
		"MaxSigners": 51,
		"VSetMask": 2251799813685247,
		"SessionKeys": {
			"PubKeys": {
				"0": "AfbUEZE7sB9guJ1dzUTN8W+Lpw8zub3vvK8Em7+3660=",
				"1": "wOcpkV6x3/vQHK7NaPiX1czKYaBpQSTAZ/BJoGRClFQ=",
				"10": "bhe8hZs0z/BckHOiHpv0m3Yvl7/zWjOxXwV9ROdZeEQ="
			}
		},
		"R1": {
			"Mask": 2251799813685247,
			"Count": 51
		},
		"R2": {
			"Mask": 2251799813685247,
			"Count": 51
		},
		"R3": {
			"Mask": 2251799813685247,
			"Count": 51,
			"Data": {
				"PubkeyPackage": "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",
				"InternalKey": "r10oJfTICETAWU7PqE6dy+49T25OT/y+7G6lbXAyDCM="
			}
		},
		"Claims": {
			"Mask": 0,
			"Count": 0,
			"Counters": {}
		},
		"CfgHash": "kZmmgWPEC6lKPRCWCo+EOwK61f/LkFonorspbmGrVg4=",
		"Attempts": 0,
		"Until": "2025-08-21T15:24:47Z"
	}`

	m := MustUnmarshalJSONMap(t, js)

	got, err := DeserializeDkg(m)
	if err != nil {
		t.Fatalf("DeserializeDkg error: %v", err)
	}

	// Simple scalars
	if got.State != coordinator.DKGStateFinished {
		t.Errorf("State=%d, want %d", got.State, coordinator.DKGStateFinished)
	}
	if got.MaxSigners != 51 {
		t.Errorf("MaxSigners=%d, want 51", got.MaxSigners)
	}
	if got.Attempts != 0 {
		t.Errorf("Attempts=%d, want 0", got.Attempts)
	}

	// VSetMask (big.Int)
	wantMask := big.NewInt(0).SetUint64(2251799813685247)
	if got.VSetMask == nil || got.VSetMask.Cmp(wantMask) != 0 {
		t.Errorf("VSetMask=%v, want %v", got.VSetMask, wantMask)
	}

	// VSet (map[uint16][]byte)
	if len(got.VSet) != 4 {
		t.Fatalf("VSet len=%d, want 4", len(got.VSet))
	}
	wantV0, _ := base64.StdEncoding.DecodeString("x4UOi03GzPdmCyhlvnT9MGINIXctTdR4brnf6yIReuQ=")
	if !bytes.Equal(got.VSet[0], wantV0) {
		t.Errorf("VSet[0] mismatch: %x vs %x", got.VSet[0], wantV0)
	}
	wantV11, _ := base64.StdEncoding.DecodeString("9XGt1LzAEORqoVUFgUvZGnjyzYuqymfRMu8G3nzhTLw=")
	if !bytes.Equal(got.VSet[11], wantV11) {
		t.Errorf("VSet[11] mismatch")
	}

	// SessionKeys
	if got.SessionKeys == nil {
		t.Fatalf("SessionKeys is nil, want non-nil")
	}
	if len(got.SessionKeys.PubKeys) != 3 {
		t.Fatalf("SessionKeys.PubKeys len=%d, want 3", len(got.SessionKeys.PubKeys))
	}
	wantPK10, _ := base64.StdEncoding.DecodeString("bhe8hZs0z/BckHOiHpv0m3Yvl7/zWjOxXwV9ROdZeEQ=")
	if !bytes.Equal(got.SessionKeys.PubKeys[10], wantPK10) {
		t.Errorf("SessionKeys.PubKeys[10] mismatch")
	}

	// R1
	if got.R1 == nil {
		t.Fatalf("R1 is nil, want non-nil")
	}
	if got.R1.Count != 51 {
		t.Errorf("R1.Count=%d, want 51", got.R1.Count)
	}
	if got.R1.Mask == nil || got.R1.Mask.Cmp(wantMask) != 0 {
		t.Errorf("R1.Mask=%v, want %v", got.R1.Mask, wantMask)
	}

	// R2
	if got.R2 == nil {
		t.Fatalf("R2 is nil, want non-nil")
	}
	if got.R2.Count != 51 {
		t.Errorf("R2.Count=%d, want 51", got.R2.Count)
	}
	if got.R2.Mask == nil || got.R2.Mask.Cmp(wantMask) != 0 {
		t.Errorf("R2.Mask=%v, want %v", got.R2.Mask, wantMask)
	}

	// R3
	if got.R3 == nil {
		t.Fatalf("R3 is nil, want non-nil")
	}
	if got.R3.Count != 51 {
		t.Errorf("R3.Count=%d, want 51", got.R3.Count)
	}
	if got.R3.Mask == nil || got.R3.Mask.Cmp(wantMask) != 0 {
		t.Errorf("R3.Mask=%v, want %v", got.R3.Mask, wantMask)
	}
	if got.R3.Data == nil {
		t.Fatalf("R3.Data is nil, want non-nil")
	}
	if len(got.R3.Data.PubkeyPackage) == 0 {
		t.Errorf("R3.Data.PubkeyPackage empty, want non-empty")
	}
	wantInternal, _ := base64.StdEncoding.DecodeString("r10oJfTICETAWU7PqE6dy+49T25OT/y+7G6lbXAyDCM=")
	if !bytes.Equal(got.R3.Data.InternalKey, wantInternal) {
		t.Errorf("R3.Data.InternalKey mismatch")
	}

	// Claims
	if got.Claims == nil {
		t.Fatalf("Claims is nil, want non-nil (present in JSON)")
	}
	if got.Claims.Mask == nil || got.Claims.Mask.Sign() != 0 {
		t.Errorf("Claims.Mask=%v, want 0", got.Claims.Mask)
	}
	if got.Claims.Count != 0 {
		t.Errorf("Claims.Count=%d, want 0", got.Claims.Count)
	}
	if len(got.Claims.Counters) != 0 {
		t.Errorf("Claims.Counters len=%d, want 0", len(got.Claims.Counters))
	}

	// CfgHash
	wantCfg, _ := base64.StdEncoding.DecodeString("kZmmgWPEC6lKPRCWCo+EOwK61f/LkFonorspbmGrVg4=")
	if !bytes.Equal(got.CfgHash, wantCfg) {
		t.Errorf("CfgHash mismatch")
	}

	// Until (RFC3339)
	wantUntil := time.Date(2025, 8, 21, 15, 24, 47, 0, time.UTC)
	if !got.Until.Equal(wantUntil) {
		t.Errorf("Until=%v, want %v", got.Until, wantUntil)
	}
}

func TestDeserializeDkg_EmptyOK(t *testing.T) {
	m := map[string]interface{}{}

	got, err := DeserializeDkg(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// zero / nil checks
	if got.State != coordinator.DKGState(0) {
		t.Errorf("default State=%d, want 0", got.State)
	}
	if len(got.VSet) != 0 {
		t.Errorf("`VSet`=%v, want nil/empty", got.VSet)
	}
	if got.MaxSigners != 0 {
		t.Errorf("MaxSigners=%d, want 0", got.MaxSigners)
	}
	if got.VSetMask != nil {
		t.Errorf("VSetMask should be nil by default")
	}
	if got.SessionKeys != nil {
		t.Errorf("SessionKeys should be nil by default")
	}
	if got.R1 != nil || got.R2 != nil || got.R3 != nil {
		t.Errorf("R1/R2/R3 should be nil by default")
	}
	if got.Claims != nil {
		t.Errorf("Claims should be nil by default")
	}
	if got.CfgHash != nil {
		t.Errorf("CfgHash should be nil by default")
	}
	if !got.Until.IsZero() {
		t.Errorf("Until should be zero time by default")
	}
}

func TestDeserializeDkg_Errors(t *testing.T) {
	t.Run("State bad type", func(t *testing.T) {
		m := map[string]interface{}{"State": "oops"}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`State` parse error") {
			t.Fatalf("want State parse error, got: %v", err)
		}
	})

	t.Run("VSet wrong type", func(t *testing.T) {
		m := map[string]interface{}{"VSet": 123}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`VSet` parse error") {
			t.Fatalf("want VSet parse error, got: %v", err)
		}
	})

	t.Run("VSetMask wrong type", func(t *testing.T) {
		m := map[string]interface{}{"VSetMask": "not-a-number"}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`VSetMask` parse error") {
			t.Fatalf("want VSetMask parse error, got: %v", err)
		}
	})

	t.Run("SessionKeys PubKeys wrong type", func(t *testing.T) {
		m := map[string]interface{}{
			"SessionKeys": map[string]interface{}{
				"PubKeys": 17,
			},
		}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`SessionKeys:PubKeys` parse error") {
			t.Fatalf("want SessionKeys:PubKeys parse error, got: %v", err)
		}
	})

	t.Run("R1 nested error (Count wrong type)", func(t *testing.T) {
		m := map[string]interface{}{
			"R1": map[string]interface{}{
				"Mask":  1,
				"Count": "nope",
			},
		}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`R1` parse error") {
			t.Fatalf("want R1 parse error, got: %v", err)
		}
	})

	t.Run("CfgHash bad base64", func(t *testing.T) {
		m := map[string]interface{}{"CfgHash": "!!notb64!!"}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`CfgHash` parse error") {
			t.Fatalf("want CfgHash parse error, got: %v", err)
		}
	})

	t.Run("Attempts wrong type", func(t *testing.T) {
		m := map[string]interface{}{"Attempts": "zero"}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`Attempts` parse error") {
			t.Fatalf("want Attempts parse error, got: %v", err)
		}
	})

	t.Run("Until bad format", func(t *testing.T) {
		m := map[string]interface{}{"Until": "21-08-2025 15:24:47"}
		_, err := DeserializeDkg(m)
		if err == nil || !strings.Contains(err.Error(), "`Until` parse error") {
			t.Fatalf("want Until parse error, got: %v", err)
		}
	})
}
