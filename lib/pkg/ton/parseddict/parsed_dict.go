package parseddict

import (
	"errors"
	"fmt"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

func NewFromDictCell[V any](dictCell *cell.Dictionary, parseKey func(*cell.Slice, uint) string, parseValue func(*cell.Slice) (V, error)) (*map[string]V, error) {
	if dictCell == nil {
		return nil, errors.New("dict cell is nil")
	}

	dictKV, err := dictCell.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load dictionary: %w", err)
	}

	result := make(map[string]V)
	for _, kv := range dictKV {
		key := parseKey(kv.Key, dictCell.GetKeySize())

		value, err := parseValue(kv.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value for key %s: %w", key, err)
		}

		result[key] = value
	}

	return &result, nil
}
