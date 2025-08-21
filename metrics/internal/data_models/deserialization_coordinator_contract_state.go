package data_models

import (
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/xssnick/tonutils-go/address"
)

func DeserializeCoordinatorContractState(json map[string]interface{}) (*coordinator.Storage, error) {
	var storage coordinator.Storage

	// Initiated (bool)
	if v, ok := json["Initiated"]; ok {
		b, err := ToBool(v)
		if err != nil {
			return nil, fmt.Errorf("`Initiated` parse error: %w", err)
		}
		storage.Initiated = b
	}

	// StandaloneMode (bool)
	if v, ok := json["StandaloneMode"]; ok {
		b, err := ToBool(v)
		if err != nil {
			return nil, fmt.Errorf("`StandaloneMode` parse error: %w", err)
		}
		storage.StandaloneMode = b
	}

	// Id (uint32)
	if v, ok := json["Id"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`Id` parse error: %w", err)
		}
		storage.Id = uint32(u)
	}

	// ConfiguratorAddr (*address.Address)
	if v, ok := json["ConfiguratorAddr"]; ok {
		addrStr, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`ConfiguratorAddr` parse error: %w", err)
		}
		a, err := address.ParseRawAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("`ConfiguratorAddr` parse error: %w", err)
		}
		storage.ConfiguratorAddr = a
	}

	// Enabled (bool)
	if v, ok := json["Enabled"]; ok {
		b, err := ToBool(v)
		if err != nil {
			return nil, fmt.Errorf("`Enabled` parse error: %w", err)
		}
		storage.Enabled = b
	}

	// UnsignedPegouts []PegoutRecord
	if v, ok := json["UnsignedPegouts"].([]interface{}); ok {
		if len(v) > 0 {
			a, err := DeserializePegouts(v, -1)
			if err != nil {
				return nil, fmt.Errorf("`UnsignedPegouts` parse error: %w", err)
			}
			storage.UnsignedPegouts = a
		} else {
			storage.UnsignedPegouts = make([]coordinator.PegoutRecord, 0)
		}
	}

	// MinClaimsPercent (uint16)
	if v, ok := json["MinClaimsPercent"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`MinClaimsPercent` parse error: %w", err)
		}
		storage.MinClaimsPercent = uint16(u)
	}

	// MinSignersThreshold (uint16)
	if v, ok := json["MinSignersThreshold"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`MinSignersThreshold` parse error: %w", err)
		}
		storage.MinSignersThreshold = uint16(u)
	}

	// DkgLifetime (uint32)
	if v, ok := json["DkgLifetime"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`DkgLifetime` parse error: %w", err)
		}
		storage.DkgLifetime = uint32(u)
	}

	// SigningTimeout (uint32)
	if v, ok := json["SigningTimeout"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`SigningTimeout` parse error: %w", err)
		}
		storage.SigningTimeout = uint32(u)
	}

	// NextPegoutIdx (uint64)
	if v, ok := json["NextPegoutIdx"]; ok {
		u, err := ToUint64(v)
		if err != nil {
			return nil, fmt.Errorf("`NextPegoutIdx` parse error: %w", err)
		}
		storage.NextPegoutIdx = uint64(u)
	}

	// TeleportAddr (*address.Address)
	if v, ok := json["TeleportAddr"]; ok {
		addrStr, err := ToString(v)
		if err != nil {
			return nil, fmt.Errorf("`TeleportAddr` parse error: %w", err)
		}
		a, err := address.ParseRawAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("`TeleportAddr` parse error: %w", err)
		}
		storage.TeleportAddr = a
	}

	return &storage, nil
}
