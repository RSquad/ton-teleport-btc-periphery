package data_models

import (
	"encoding/json"
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializeCoordinatorContractStorage(jsonData []byte) (*coordinator.Storage, error) {
	if jsonData == nil {
		return nil, errors.New("jsonData is null")
	}

	var storage coordinator.Storage
	err := json.Unmarshal(jsonData, &storage)
	if err != nil {
		return nil, err
	}

	return &storage, nil
}
