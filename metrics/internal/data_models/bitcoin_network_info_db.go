package data_models

import (
	"encoding/json"
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
		return nil, err
	}

	return &info, nil
}

func SerializeBitcoinNetworkInfoDB(info *BitcoinNetworkInfo) ([]byte, error) {
	return json.Marshal(info)
}
