package data_models

import (
	"encoding/json"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

func DeserializeTeleportContractStorage(jsonData []byte) (*teleportcontract.Storage, error) {
	var storage teleportcontract.Storage
	err := json.Unmarshal(jsonData, &storage)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}

func SerializeTeleportContractStorage(storage *teleportcontract.Storage) ([]byte, error) {
	return json.Marshal(storage)
}
