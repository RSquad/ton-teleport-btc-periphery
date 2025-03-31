package parseddict

import (
	"errors"
	"fmt"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

//func ParseKey(s *cell.Slice, keySize uint) string {
//	str := fmt.Sprintf("%x", s.MustLoadBigUInt(keySize).Bytes())
//	return fmt.Sprintf("%0*s", ((keySize+7)/8)*2, str)
//}

func ParseKey(s *cell.Slice, keySize uint) string {
	return fmt.Sprintf("%x", s.MustLoadBigUInt(keySize).FillBytes(make([]byte, (keySize+7)/8)))
}

//func ParseKey(s *cell.Slice, keySize uint) string {
//	return fmt.Sprintf("%x", s.MustLoadBigUInt(keySize).FillBytes(make([]byte, (keySize+7)/8)))
//}
//000000000000000000000000000000fa

func New[V any](
	dict *cell.Dictionary,
	parseKey func(*cell.Slice, uint) string,
	parseValue func(*cell.Slice) (V, error),
) (*map[string]V, error) {
	return ParseDict(dict, parseKey, parseValue)
}

func ParseDict[K comparable, V any](
	dict *cell.Dictionary,
	parseKey func(*cell.Slice, uint) K,
	parseValue func(*cell.Slice) (V, error),
) (*map[K]V, error) {
	if dict == nil {
		return nil, errors.New("dict cell is nil")
	}

	dictKV, err := dict.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load dictionary: %w", err)
	}

	result := make(map[K]V)
	for _, kv := range dictKV {
		key := parseKey(kv.Key, dict.GetKeySize())

		value, err := parseValue(kv.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value for key %w", err)
		}

		result[key] = value
	}

	return &result, nil
}
