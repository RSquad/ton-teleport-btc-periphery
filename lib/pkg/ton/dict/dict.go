package dict

import (
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Parser interface {
	ParseKey(*cell.Slice)
	ParseValue(*cell.Slice)
}

type Dictionary interface {
	Get()
	GetByKey(key any)
}

type Dict[K comparable, V any] struct {
	parseKey       func(*cell.Slice) K
	parseValue     func(*cell.Slice) V
	cellDictionary *cell.Dictionary
	dictionary     map[K]V
}

func (p *Dict[K, V]) Parse() map[K]V {
	dictKV, err := p.cellDictionary.LoadAll()
	if err != nil {
		return nil
	}

	dict := make(map[K]V, len(dictKV))

	for _, kv := range dictKV {
		key := p.parseKey(kv.Key)
		value := p.parseValue(kv.Value)

		dict[key] = value
	}
	p.dictionary = dict
	return p.dictionary
}

func (p *Dict[K, V]) Get() map[K]V {
	return p.dictionary
}

func (p *Dict[K, V]) GetByKey(key K) V {
	return p.dictionary[key]
}

func WriteCellsToBuffer(cell *cell.Slice) [][]byte {
	var cellBytes [][]byte

	loadedBytes, _ := cell.LoadSlice(cell.BitsLeft() / 8)
	cellBytes = append(cellBytes, loadedBytes)
	for i := 1; i < cell.RefsNum()+1; i++ {
		cellSlice := cell.MustLoadRef()
		loadedBytes, _ := cellSlice.LoadSlice(cellSlice.BitsLeft() / 8)
		cellBytes = append(cellBytes, loadedBytes)
	}
	return cellBytes
}
