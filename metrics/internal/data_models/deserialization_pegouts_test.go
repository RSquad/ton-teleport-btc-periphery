package data_models

import (
	"encoding/base64"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/xssnick/tonutils-go/address"
)

func TestDeserializePegouts_Success(t *testing.T) {
	jsonInput := `
[
  {
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
  }
]`

	arr := MustUnmarshalJSONArray(t, jsonInput)

	got, err := DeserializePegouts(arr, 10)
	if err != nil {
		t.Fatalf("DeserializePegouts error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got)=%d, want 1", len(got))
	}

	pegout := got[0]

	// ID
	if pegout.ID != 135 {
		t.Errorf("ID=%d, want 135", pegout.ID)
	}

	// Address
	wantAddr, _ := address.ParseAddr("EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-")
	if pegout.PegoutAddress == nil || pegout.PegoutAddress.String() != wantAddr.String() {
		t.Errorf("PegoutAddress=%v, want %v", pegout.PegoutAddress, wantAddr)
	}

	// InternalKey
	wantKey, _ := base64.StdEncoding.DecodeString("ZwqZwANOMRDlrCsHEsOeem9vDFljXk5IDh4H+OMhDTk=")
	if !reflect.DeepEqual(pegout.InternalKey, wantKey) {
		t.Errorf("InternalKey mismatch")
	}

	// IsAutopegout
	if !pegout.IsAutopegout {
		t.Errorf("IsAutopegout=false, want true")
	}

	// Commitments
	if len(pegout.Commitments) != 2 {
		t.Fatalf("Commitments len=%d, want 2", len(pegout.Commitments))
	}
	val46, _ := base64.StdEncoding.DecodeString("AQAjD4qzA+NRZBmoOjp0ftul6qYqcwR147bphPYltmE5VLIYCxcOAz18T35KTKHp0+O5RHk6kI+ni5qHIw3Mkf6PZgZ1wFoB")
	if !reflect.DeepEqual(pegout.Commitments[46], val46) {
		t.Errorf("Commitments[46] mismatch")
	}

	// Big.Int masks
	if pegout.CommitmentsMaskAccepted.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("CommitmentsMaskAccepted != 0")
	}
	wantOther, _ := new(big.Int).SetString("1237940039285450643643301888", 10)
	if pegout.CommitmentsMaskOther.Cmp(wantOther) != 0 {
		t.Errorf("CommitmentsMaskOther=%v, want %v", pegout.CommitmentsMaskOther, wantOther)
	}

	// SigningSharesMask
	emptyMask, _ := base64.StdEncoding.DecodeString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	if !reflect.DeepEqual(pegout.SigningSharesMask, emptyMask) {
		t.Errorf("SigningSharesMask mismatch")
	}

	// Signatures
	if pegout.Signatures.Mask == nil || pegout.Signatures.Mask.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Signatures.Mask != 0")
	}
	if pegout.Signatures.Count != 0 {
		t.Errorf("Signatures.Count=%d, want 0", pegout.Signatures.Count)
	}
	if !reflect.DeepEqual(pegout.Signatures.Hash, emptyMask) {
		t.Errorf("Signatures.Hash mismatch")
	}

	// Claims
	if pegout.ClaimsMask == nil || pegout.ClaimsMask.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("ClaimsMask != 0")
	}
	if pegout.ClaimsCount != 0 {
		t.Errorf("ClaimsCount=%d, want 0", pegout.ClaimsCount)
	}
	if len(pegout.ClaimsCounters) != 0 {
		t.Errorf("ClaimsCounters len=%d, want 0", len(pegout.ClaimsCounters))
	}

	// MaxSigners
	if pegout.MaxSigners != 91 {
		t.Errorf("MaxSigners=%d, want 91", pegout.MaxSigners)
	}

	// ExpiredAt
	wantTime := time.Date(2025, 7, 16, 3, 27, 21, 0, time.UTC)
	if !pegout.ExpiredAt.Equal(wantTime) {
		t.Errorf("ExpiredAt=%v, want %v", pegout.ExpiredAt, wantTime)
	}

	// SigningMask
	wantSigningMask, _ := new(big.Int).SetString("2475880078570760549798248447", 10)
	if pegout.SigningMask == nil || pegout.SigningMask.Cmp(wantSigningMask) != 0 {
		t.Errorf("SigningMask=%v, want %v", pegout.SigningMask, wantSigningMask)
	}
}

func TestDeserializePegouts_Limit(t *testing.T) {
	jsonInput := `
[
  {"ID":1},
  {"ID":2},
  {"ID":3}
]`
	arr := MustUnmarshalJSONArray(t, jsonInput)

	got, err := DeserializePegouts(arr, 2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("limit not respected: %+v", got)
	}
}

func TestDeserializePegouts_Errors(t *testing.T) {
	t.Run("element not object", func(t *testing.T) {
		arr := []interface{}{"not an object"}
		_, err := DeserializePegouts(arr, 1)
		if err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("missing ID", func(t *testing.T) {
		js := `[{"PegoutAddress":"EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-"}]`
		arr := MustUnmarshalJSONArray(t, js)
		_, err := DeserializePegouts(arr, 1)
		if err == nil {
			t.Fatalf("expected error for missing ID")
		}
	})

	t.Run("bad base64", func(t *testing.T) {
		js := `[{"ID":1, "InternalKey":"@@not-base64@@", "PegoutAddress":"EQAPtQRffHrXATHokYMFQgupunwxfTe2Main1FYFUt-8eHn-"}]`
		arr := MustUnmarshalJSONArray(t, js)
		_, err := DeserializePegouts(arr, 1)
		if err == nil {
			t.Fatalf("expected base64 error")
		}
	})
}
