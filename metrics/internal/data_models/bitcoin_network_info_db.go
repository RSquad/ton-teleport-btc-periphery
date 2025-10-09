package data_models

import (
	"encoding/json"
	"fmt"
)

type BitcoinNetworkInfo struct {
	Chain         string
	Blocks        int32
	BestBlockHash string
	MedianTime    int64
}

func DeserializeBitcoinNetworkInfoDB(jsonData []byte) (*BitcoinNetworkInfo, error) {
	var info BitcoinNetworkInfo
	err := json.Unmarshal(jsonData, &info)
	if err != nil {
		return nil, fmt.Errorf("failed to call `DeserializeBitcoinNetworkInfoDB`, json '%s': %w", string(jsonData), err)
	}

	return &info, nil
}

func SerializeBitcoinNetworkInfoDB(info *BitcoinNetworkInfo) ([]byte, error) {
	return json.Marshal(info)
}
