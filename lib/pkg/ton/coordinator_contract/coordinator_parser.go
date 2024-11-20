package coordinatorcontract

import (
	"fmt"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const Ed25519PubkeyTag = 0x8e81278a

type DkgPackageParams struct {
	mask  uint64
	count uint64
}

func DictParse[K comparable, V any](dictCell *cell.Slice, keySize uint,
	parseKey func(*cell.Slice) (K, error),
	parseValue func(*cell.Slice) (V, error),
) (map[K]V, error) {
	if dictCell == nil {
		panic("dictCell is nil")
	}

	dictionary := dictCell.MustLoadDict(keySize)
	if dictionary == nil {
		panic("failed to create dictionary from dictCell")
	}

	dict, err := dictionary.LoadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load all key-value pairs from dictionary: %w", err)
	}

	result := make(map[K]V, len(dict))

	for _, kv := range dict {
		key, err := parseKey(kv.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key: %w", err)
		}
		value, err := parseValue(kv.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value for key %v: %w", key, err)
		}

		result[key] = value
	}

	return result, nil
}

func PackageParse(dkg *cell.Slice) (DkgPackageParams, error) {
	mask, _ := dkg.LoadBigUInt(256)
	count, _ := dkg.LoadUInt(16)

	return DkgPackageParams{
		mask:  mask.Uint64(),
		count: count,
	}, nil
}

func ValidatorDescrValueParse(slice *cell.Slice) ([]byte, error) {
	tag, _ := slice.LoadUInt(8)
	if (tag &^ 0x20) != 0x53 {
		panic("Invalid Validator Descr tag")
	}
	pubkeyTag, _ := slice.LoadUInt(32)
	if pubkeyTag != Ed25519PubkeyTag {
		panic("Invalid PublicKey tag")
	}
	return slice.LoadSlice(256)
}

func packageValueParse(slice *cell.Slice) ([][]byte, error) {
	return writeCellsToBuffer(slice.MustLoadRef())
}

func packageDictionaryValueParse(slice *cell.Slice) (map[[32]byte][][]byte, error) {
	slice.MustLoadUInt(256)

	cellDict := slice.MustLoadDict(256)
	dict, _ := cellDict.LoadAll()

	res := make(map[[32]byte][][]byte)

	for _, kv := range dict {
		key := kv.Key.MustLoadSlice(256)
		value, _ := packageValueParse(kv.Value)
		res[[32]byte(key)] = value
	}
	return res, nil
}

func writeCellsToBuffer(cell *cell.Slice) ([][]byte, error) {
	// log.Printf("bitsSz = %v", cell.BitsLeft())
	// log.Printf("RefsNum = %v", cell.RefsNum())
	var cellBytes [][]byte

	loadedBytes, _ := cell.LoadSlice(cell.BitsLeft() / 8)
	cellBytes = append(cellBytes, loadedBytes)
	for i := 1; i < cell.RefsNum()+1; i++ {
		cellSlice := cell.MustLoadRef()
		loadedBytes, _ := cellSlice.LoadSlice(cellSlice.BitsLeft() / 8)
		cellBytes = append(cellBytes, loadedBytes)
	}
	return cellBytes, nil
}
