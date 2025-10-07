package data_models

import (
	"encoding/json"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializePegout(jsonData []byte) (*coordinator.PegoutRecord, error) {
	var pegout coordinator.PegoutRecord
	err := json.Unmarshal(jsonData, &pegout)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializePegout`, json '%s': %w", string(jsonData), err)
	}

	return &pegout, nil
}

func DeserializePegouts(jsonData []byte) ([]coordinator.PegoutRecord, error) {
	var pegouts []coordinator.PegoutRecord
	err := json.Unmarshal(jsonData, &pegouts)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializePegouts`, json '%s': %w", string(jsonData), err)
	}

	return pegouts, nil
}
