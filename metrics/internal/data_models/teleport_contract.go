package data_models

import (
	"encoding/json"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

func DeserializeTeleportContractStorage(jsonData []byte) (*teleportcontract.Storage, error) {
	var storage teleportcontract.Storage
	err := json.Unmarshal(jsonData, &storage)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializeTeleportContractStorage`, json '%s': %w", string(jsonData), err)
	}

	return &storage, nil
}

func SerializeTeleportContractStorage(storage *teleportcontract.Storage) ([]byte, error) {
	return json.Marshal(storage)
}
