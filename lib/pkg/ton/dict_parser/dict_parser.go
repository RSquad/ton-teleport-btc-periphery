package dict_parser

import (
	"fmt"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Parser interface {
	BuildParse()
	ParseKey(*cell.Slice)
	ParseValue(*cell.Slice)
}

type DictParser[K comparable, V any] struct {
	ParseKey   func(*cell.Slice) (K, error)
	ParseValue func(*cell.Slice) (V, error)
	dictKV     []cell.DictKV
}

func (p *DictParser[K, V]) Parse() (map[K]V, error) {
	dict := make(map[K]V, len(p.dictKV))

	for _, kv := range p.dictKV {
		key, err := p.ParseKey(kv.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}

		value, err := p.ParseValue(kv.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value for key %v: %w", key, err)
		}

		dict[key] = value
	}

	return dict, nil
}
