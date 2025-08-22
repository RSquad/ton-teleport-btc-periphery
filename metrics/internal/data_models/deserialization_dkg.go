package data_models

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializeDkg(json map[string]interface{}) (*coordinator.DKG, error) {
	var dkg coordinator.DKG

	// State (DKGState)
	if v, ok := json["State"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`State` parse error: %w", err)
		}
		dkg.State = coordinator.DKGState(u)
	}

	// VSet (VSet: map[uint16][]byte)
	if v, ok := json["VSet"]; ok && v != nil {
		m, err := MapStringToBytesByUint16Key(v)
		if err != nil {
			return nil, fmt.Errorf("`VSet` parse error: %w", err)
		}
		dkg.VSet = m
	}

	// MaxSigners (uint16)
	if v, ok := json["MaxSigners"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`MaxSigners` parse error: %w", err)
		}
		dkg.MaxSigners = uint16(u)
	}

	// VSetMask (*big.Int)
	if v, ok := json["VSetMask"]; ok {
		bi, err := ToBigInt(v)
		if err != nil {
			return nil, fmt.Errorf("`VSetMask` parse error: %w", err)
		}
		dkg.VSetMask = bi
	}

	// SessionKeys (*SessionKeys)
	if v1, ok := json["SessionKeys"].(map[string]interface{}); ok {
		if v2, ok := v1["PubKeys"]; ok {
			m, err := MapStringToBytesByUint16Key(v2)
			if err != nil {
				return nil, fmt.Errorf("`SessionKeys:PubKeys` parse error: %w", err)
			}
			dkg.SessionKeys = &coordinator.SessionKeys{
				PubKeys: m,
			}
		}
	}

	// R1 (*DKGR1)
	if v, ok := json["R1"].(map[string]interface{}); ok {
		r1, err := DeserializeDkgR1(v)
		if err != nil {
			return nil, fmt.Errorf("`R1` parse error: %w", err)
		}
		dkg.R1 = r1
	}

	// R2 (*DKGR2)
	if v, ok := json["R2"].(map[string]interface{}); ok {
		r2, err := DeserializeDkgR2(v)
		if err != nil {
			return nil, fmt.Errorf("`R2` parse error: %w", err)
		}
		dkg.R2 = r2
	}

	// R3 (*DKGR3)
	if v, ok := json["R3"].(map[string]interface{}); ok {
		r3, err := DeserializeDkgR3(v)
		if err != nil {
			return nil, fmt.Errorf("`R3` parse error: %w", err)
		}
		dkg.R3 = r3
	}

	// Claims (*DKGClaims)
	if v, ok := json["Claims"].(map[string]interface{}); ok {
		claims, err := DeserializeDkgClaims(v)
		if err != nil {
			return nil, fmt.Errorf("`Claims` parse error: %w", err)
		}
		dkg.Claims = claims
	}

	// CfgHash ([]byte, base64)
	if v, ok := json["CfgHash"]; ok {
		b64, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`CfgHash` parse error: %w", err)
		}
		bs, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("`CfgHash` parse error: %w", err)
		}
		dkg.CfgHash = bs
	}

	// Attempts (uint64)
	if v, ok := json["Attempts"]; ok {
		id64, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`Attempts` parse error: %w", err)
		}
		dkg.Attempts = id64
	}

	// Until (time.Time, RFC3339)
	if v, ok := json["Until"]; ok {
		s, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`Until` parse error: %w", err)
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, fmt.Errorf("`Until` parse error: %w", err)
		}
		dkg.Until = t
	}

	return &dkg, nil
}

func DeserializeDkgR1(json map[string]interface{}) (*coordinator.DKGR1, error) {
	var r1 coordinator.DKGR1

	// Mask (*big.Int)
	if v, ok := json["Mask"]; ok {
		bi, err := ToBigInt(v)
		if err != nil {
			return nil, fmt.Errorf("`Mask` parse error: %w", err)
		}
		r1.Mask = bi
	}

	// Count (uint64)
	if v, ok := json["Count"]; ok {
		id64, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`Count` parse error: %w", err)
		}
		r1.Count = id64
	}

	return &r1, nil
}

func DeserializeDkgR2(json map[string]interface{}) (*coordinator.DKGR2, error) {
	var r2 coordinator.DKGR2

	// Mask (*big.Int)
	if v, ok := json["Mask"]; ok {
		bi, err := ToBigInt(v)
		if err != nil {
			return nil, fmt.Errorf("`Mask` parse error: %w", err)
		}
		r2.Mask = bi
	}

	// Count (uint64)
	if v, ok := json["Count"]; ok {
		id64, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`Count` parse error: %w", err)
		}
		r2.Count = id64
	}

	return &r2, nil
}

func DeserializeDkgR3(json map[string]interface{}) (*coordinator.DKGR3, error) {
	var r3 coordinator.DKGR3

	// Mask (*big.Int)
	if v, ok := json["Mask"]; ok {
		bi, err := ToBigInt(v)
		if err != nil {
			return nil, fmt.Errorf("`Mask` parse error: %w", err)
		}
		r3.Mask = bi
	}

	// Count (uint64)
	if v, ok := json["Count"]; ok {
		id64, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`Count` parse error: %w", err)
		}
		r3.Count = uint16(id64)
	}

	// Data (*PubkeyData)
	if v, ok := json["Data"].(map[string]interface{}); ok {
		pubkeyPackage := make([]byte, 0)
		internalKey := make([]byte, 0)

		// PubkeyPackage ([]byte, base64)
		if pubkeyPackageV, ok := v["PubkeyPackage"]; ok {
			b64, err := ToString(pubkeyPackageV)
			if err != nil {
				return nil, fmt.Errorf("`PubkeyPackage` parse error: %w", err)
			}
			bs, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("`PubkeyPackage` parse error: %w", err)
			}
			pubkeyPackage = bs
		}

		// InternalKey ([]byte)
		if internalKeyV, ok := v["InternalKey"]; ok {
			b64, err := ToString(internalKeyV)
			if err != nil {
				return nil, fmt.Errorf("`InternalKey` parse error: %w", err)
			}
			bs, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("`InternalKey` parse error: %w", err)
			}
			internalKey = bs
		}

		r3.Data = &coordinator.PubkeyData{
			PubkeyPackage: pubkeyPackage,
			InternalKey:   internalKey,
		}
	}

	return &r3, nil
}

func DeserializeDkgClaims(json map[string]interface{}) (*coordinator.DKGClaims, error) {
	var claims coordinator.DKGClaims

	// Mask (*big.Int)
	if v, ok := json["Mask"]; ok {
		bi, err := ToBigInt(v)
		if err != nil {
			return nil, fmt.Errorf("`Mask` parse error: %w", err)
		}
		claims.Mask = bi
	}

	// Count (uint16)
	if v, ok := json["Count"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`Count` parse error: %w", err)
		}
		claims.Count = uint16(u)
	}

	// Counters (DKGClaimcounters: map[uint16]uint16)
	if v, ok := json["Counters"].(map[string]interface{}); ok {
		counters, err := MapStringToUint16ByUint16Key(v)
		if err != nil {
			return nil, fmt.Errorf("`Counters` parse error: %w", err)
		}
		claims.Counters = counters
	}

	return &claims, nil
}
