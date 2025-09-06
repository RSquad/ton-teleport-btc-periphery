package data_models

import (
	"encoding/json"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializeCoordinatorContractStorage(jsonData []byte) (*coordinator.Storage, error) {
	var storage coordinator.Storage
	err := json.Unmarshal(jsonData, &storage)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}
