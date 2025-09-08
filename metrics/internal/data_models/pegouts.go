package data_models

import (
	"encoding/json"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
)

func DeserializePegout(jsonData []byte) (*coordinator.PegoutRecord, error) {
	var pegout coordinator.PegoutRecord
	err := json.Unmarshal(jsonData, &pegout)
	if err != nil {
		return nil, err
	}

	return &pegout, nil
}

func DeserializePegouts(jsonData []byte) ([]coordinator.PegoutRecord, error) {
	var pegouts []coordinator.PegoutRecord
	err := json.Unmarshal(jsonData, &pegouts)
	if err != nil {
		return nil, err
	}

	return pegouts, nil
}
