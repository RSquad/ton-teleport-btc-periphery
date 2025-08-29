package data_models

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/xssnick/tonutils-go/address"
)

func DeserializePegouts(json []interface{}, limit int) ([]coordinator.PegoutRecord, error) {
	if limit < 0 || limit > len(json) {
		limit = len(json)
	}

	if len(json) == 0 {
		return nil, errors.New("json is empty")
	}

	pegouts := make([]coordinator.PegoutRecord, 0, limit)

	for i := 0; i < limit; i++ {
		raw, ok := json[i].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("element %d is not an object", i)
		}

		var pegout coordinator.PegoutRecord

		// ID (uint64)
		if v, ok := raw["ID"]; ok {
			id64, err := ToUint64(v)
			if err != nil {
				return nil, fmt.Errorf("ID at %d: %w", i, err)
			}
			pegout.ID = id64
		}

		// PegoutAddress (*address.Address)
		if v, ok := raw["PegoutAddress"]; ok {
			addrStr, err := ToString(v)
			if err != nil {
				return nil, fmt.Errorf("PegoutAddress at %d: %w", i, err)
			}
			a, err := address.ParseAddr(addrStr)
			if err != nil {
				return nil, fmt.Errorf("PegoutAddress parse at %d: %w", i, err)
			}
			pegout.PegoutAddress = a
		}

		// InternalKey ([]byte, base64)
		if v, ok := raw["InternalKey"]; ok {
			b64, err := ToString(v)
			if err != nil {
				return nil, fmt.Errorf("InternalKey at %d: %w", i, err)
			}
			bs, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("InternalKey decode at %d: %w", i, err)
			}
			pegout.InternalKey = bs
		}

		// IsAutopegout (bool)
		if v, ok := raw["IsAutopegout"]; ok {
			b, err := ToBool(v)
			if err != nil {
				return nil, fmt.Errorf("IsAutopegout at %d: %w", i, err)
			}
			pegout.IsAutopegout = b
		}

		// Commitments (map[uint16][]byte) values are base64 strings
		if v, ok := raw["Commitments"]; ok && v != nil {
			m, err := MapStringToBytesByUint16Key(v)
			if err != nil {
				return nil, fmt.Errorf("commitments at %d: %w", i, err)
			}
			pegout.Commitments = m
		}

		// CommitmentsMaskAccepted (*big.Int)
		if v, ok := raw["CommitmentsMaskAccepted"]; ok {
			bi, err := ToBigInt(v)
			if err != nil {
				return nil, fmt.Errorf("CommitmentsMaskAccepted at %d: %w", i, err)
			}
			pegout.CommitmentsMaskAccepted = bi
		}

		// CommitmentsMaskOther (*big.Int)
		if v, ok := raw["CommitmentsMaskOther"]; ok {
			bi, err := ToBigInt(v)
			if err != nil {
				return nil, fmt.Errorf("CommitmentsMaskOther at %d: %w", i, err)
			}
			pegout.CommitmentsMaskOther = bi
		}

		// SigningShares (map[uint16]map[uint16][]byte) base64 values
		if v, ok := raw["SigningShares"]; ok && v != nil {
			m, err := Map2DStringToBytesByUint16Keys(v)
			if err != nil {
				return nil, fmt.Errorf("SigningShares at %d: %w", i, err)
			}
			pegout.SigningShares = m
		}

		// SigningSharesMask ([]byte, base64)
		if v, ok := raw["SigningSharesMask"]; ok {
			b64, err := ToString(v)
			if err != nil {
				return nil, fmt.Errorf("SigningSharesMask at %d: %w", i, err)
			}
			bs, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				return nil, fmt.Errorf("SigningSharesMask decode at %d: %w", i, err)
			}
			pegout.SigningSharesMask = bs
		}

		// Signatures (coordinator.PegoutSignatures) — customize fillSignatures if your type differs
		if v, ok := raw["Signatures"]; ok && v != nil {
			sig, err := ParsePegoutSignatures(v)
			if err != nil {
				return nil, fmt.Errorf("signatures at %d: %w", i, err)
			}

			pegout.Signatures = *sig
		}

		// ClaimsMask (*big.Int)
		if v, ok := raw["ClaimsMask"]; ok {
			bi, err := ToBigInt(v)
			if err != nil {
				return nil, fmt.Errorf("ClaimsMask at %d: %w", i, err)
			}
			pegout.ClaimsMask = bi
		}

		// ClaimsCount (uint16)
		if v, ok := raw["ClaimsCount"]; ok {
			u, err := ToUint64(v)
			if err != nil {
				return nil, fmt.Errorf("ClaimsCount at %d: %w", i, err)
			}
			pegout.ClaimsCount = uint16(u)
		}

		// ClaimsCounters (map[uint16]uint16)
		if v, ok := raw["ClaimsCounters"]; ok && v != nil {
			m, err := MapStringToUint16ByUint16Key(v)
			if err != nil {
				return nil, fmt.Errorf("ClaimsCounters at %d: %w", i, err)
			}
			pegout.ClaimsCounters = m
		}

		// MaxSigners (uint16)
		if v, ok := raw["MaxSigners"]; ok {
			u, err := ToUint64(v)
			if err != nil {
				return nil, fmt.Errorf("MaxSigners at %d: %w", i, err)
			}
			pegout.MaxSigners = uint16(u)
		}

		// ExpiredAt (time.Time, RFC3339)
		if v, ok := raw["ExpiredAt"]; ok {
			s, err := ToString(v)
			if err != nil {
				return nil, fmt.Errorf("ExpiredAt at %d: %w", i, err)
			}
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return nil, fmt.Errorf("ExpiredAt parse at %d: %w", i, err)
			}
			pegout.ExpiredAt = t
		}

		// SigningMask (*big.Int)
		if v, ok := raw["SigningMask"]; ok {
			bi, err := ToBigInt(v)
			if err != nil {
				return nil, fmt.Errorf("SigningMask at %d: %w", i, err)
			}
			pegout.SigningMask = bi
		}

		pegouts = append(pegouts, pegout)
	}

	return pegouts, nil
}

func ParsePegoutSignatures(v interface{}) (*coordinator.PegoutSignatures, error) {
	obj, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("signatures is not an object")
	}

	var mask *big.Int
	if mv, ok := obj["Mask"]; ok {
		bi, err := ToBigInt(mv)
		if err != nil {
			return nil, fmt.Errorf("signatures.Mask: %w", err)
		}
		mask = bi
	}

	var count uint16
	if cv, ok := obj["Count"]; ok {
		u, err := ToUint64(cv)
		if err != nil {
			return nil, fmt.Errorf("signatures.Count: %w", err)
		}
		count = uint16(u)
	}

	var hash []byte
	if hv, ok := obj["Hash"]; ok {
		s, err := ToString(hv)
		if err != nil {
			return nil, fmt.Errorf("signatures.Hash: %w", err)
		}
		bs, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("signatures.Hash decode: %w", err)
		}
		hash = bs
	}

	return &coordinator.PegoutSignatures{
		Mask:  mask,
		Count: count,
		Hash:  hash,
	}, nil
}

func DeserializePegoutDbRow(json map[string]interface{}) (*PegoutDbRow, error) {
	var pegoutDbRow PegoutDbRow

	// Id
	if v, ok := json["id"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`id` parse error: %w", err)
		}
		pegoutDbRow.Id = u
	}

	// Addr
	if v, ok := json["addr"]; ok {
		s, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`addr` parse error: %w", err)
		}

		addr, err := address.ParseAddr(s)
		if err != nil {
			return nil, err
		}
		pegoutDbRow.Addr = addr
	}

	// Status
	if v, ok := json["status"]; ok {
		s, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`status` parse error: %w", err)
		}

		status, err := PegoutStatusFromString(s)
		if err != nil {
			return nil, fmt.Errorf("`status` parse error: %w", err)
		}

		pegoutDbRow.Status = status
	}

	// BitcoinTxRaw
	if v, ok := json["bitcoin_tx_raw"]; ok {
		s, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`bitcoin_tx_raw` parse error: %w", err)
		}

		var data []byte = nil
		if len(s) > 0 {
			data, err = hex.DecodeString(s)
			if err != nil {
				return nil, fmt.Errorf("`bitcoin_tx_raw` parse error: %w", err)
			}
		}

		pegoutDbRow.BitcoinTxRaw = data
	}

	// BitcoinTxId
	if v, ok := json["bitcoin_tx_id"]; ok {
		s, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`bitcoin_tx_id` parse error: %w", err)
		}

		var data []byte = nil
		if len(s) > 0 {
			data, err = hex.DecodeString(s)
			if err != nil {
				return nil, fmt.Errorf("`bitcoin_tx_id` parse error: %w", err)
			}
		}

		pegoutDbRow.BitcoinTxId = data
	}

	// BitcoinBlockHash
	if v, ok := json["bitcoin_block_hash"]; ok {
		s, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`bitcoin_block_hash` parse error: %w", err)
		}

		var data []byte = nil
		if len(s) > 0 {
			data, err = hex.DecodeString(s)
			if err != nil {
				return nil, fmt.Errorf("`bitcoin_block_hash` parse error: %w", err)
			}
		}

		pegoutDbRow.BitcoinBlockHash = data
	}

	return &pegoutDbRow, nil
}
