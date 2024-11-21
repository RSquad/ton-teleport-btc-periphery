package dict

import "github.com/xssnick/tonutils-go/tvm/cell"

type (
	R2PackageKey   [32]byte
	R2PackageValue map[[32]byte][][]byte
)

type R2PackageDict struct {
	Dict[R2PackageKey, R2PackageValue]
}

func (b *R2PackageDict) NewDict(cellDictionary *cell.Dictionary) *Dict[R2PackageKey, R2PackageValue] {
	dict := &Dict[R2PackageKey, R2PackageValue]{
		parseKey:       b.parseKey,
		parseValue:     b.parseValue,
		cellDictionary: cellDictionary,
	}
	return dict
}

func (b *R2PackageDict) parseKey(key *cell.Slice) R2PackageKey {
	k := key.MustLoadSlice(256)

	return R2PackageKey(k)
}

func (b *R2PackageDict) parseValue(value *cell.Slice) R2PackageValue {
	value.MustLoadUInt(256)
	cellDict := value.MustLoadDict(256)
	dict, _ := cellDict.LoadAll()

	res := make(map[[32]byte][][]byte)

	for _, kv := range dict {
		key := kv.Key.MustLoadSlice(256)
		res[R2PackageKey(key)] = writeCellsToBuffer(kv.Value.MustLoadRef())
	}

	return res
}
