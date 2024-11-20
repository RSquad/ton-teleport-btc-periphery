package coordinatorcontract

import (
	"fmt"

	"github.com/xssnick/tonutils-go/address"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const Ed25519PubkeyTag = 0x8e81278a

type DkgPackageParams struct {
	mask  uint64
	count uint64
}

type TPegoutRecord struct {
	internalKey    [32]byte
	pegoutAddress  *address.Address
	commitments    map[[32]byte][][]byte
	signingShares  map[[32]byte]*cell.Cell
	commitmentMask []byte
	signSharesMask []byte
}

func dictParse[K comparable, V any](dictCell *cell.Slice, keySize uint,
	parseKey func(*cell.Slice) (K, error),
	parseValue func(*cell.Slice) (V, error),
) (map[K]V, error) {
	dictionary := dictCell.MustLoadDict(keySize)
	if dictionary == nil {
		return nil, fmt.Errorf("failed to create dictionary from dictCell")
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

func loadUintKey(key *cell.Slice) (uint64, error) {
	return key.LoadUInt(16)
}

func loadBytesKey(key *cell.Slice) ([32]byte, error) {
	k, err := key.LoadSlice(256)

	return [32]byte(k), err
}

func loadCellFromValue(value *cell.Slice) (*cell.Cell, error) {
	return value.ToCell()
}

func loadPegoutRecordKey(key *cell.Slice) (uint64, error) {
	return key.LoadUInt(64)
}

func loadPegoutRecordValue(value *cell.Slice) (TPegoutRecord, error) {
	slice := value.MustLoadRef()
	commitmentMask := slice.MustLoadSlice(32)
	slice.MustLoadUInt(16)
	commitmentsDict, _ := dictParse[[32]byte, [][]byte](slice, 16, loadBytesKey, packageValueParse)

	signSharesMask := slice.MustLoadSlice(32)
	slice.MustLoadUInt(16)

	signingSharesDict, _ := dictParse[[32]byte, *cell.Cell](slice, 16, loadBytesKey, loadCellFromValue)
	pegoutAddress := slice.MustLoadAddr()
	internalKey := slice.MustLoadRef().MustLoadSlice(256)

	return TPegoutRecord{
		commitmentMask: commitmentMask,
		commitments:    commitmentsDict,
		signSharesMask: signSharesMask,
		signingShares:  signingSharesDict,
		pegoutAddress:  pegoutAddress,
		internalKey:    [32]byte(internalKey),
	}, nil
}

func packageParse(dkg *cell.Slice) (DkgPackageParams, error) {
	mask, _ := dkg.LoadBigUInt(256)
	count, _ := dkg.LoadUInt(16)

	return DkgPackageParams{
		mask:  mask.Uint64(),
		count: count,
	}, nil
}

func validatorDescrValueParse(slice *cell.Slice) ([]byte, error) {
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
