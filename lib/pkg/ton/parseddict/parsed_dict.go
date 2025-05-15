package parseddict

import (
	"errors"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func ParseKeyStr(s *cell.Slice, keySize uint) string {
	return fmt.Sprintf("%x", s.MustLoadBigUInt(keySize).FillBytes(make([]byte, (keySize+7)/8)))
}

func ParseKeyUI16(s *cell.Slice, keySize uint) uint16 {
	return uint16(s.MustLoadBigUInt(keySize).Int64())
}

func ParseKeyUI64(s *cell.Slice, keySize uint) uint64 {
	return uint64(s.MustLoadBigUInt(keySize).Int64())
}

func ParseKeyBitcoinHashAsStr(s *cell.Slice, keySize uint) string {
	hash, err := chainhash.NewHash(s.MustLoadSlice(256))
	if err != nil {
		panic(err)
	}

	return hash.String()
}

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
