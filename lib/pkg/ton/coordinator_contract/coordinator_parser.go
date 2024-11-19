package coordinatorcontract

import (
	"log"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

const Ed25519PubkeyTag = 0x8e81278a

type DkgPackageParams struct {
	mask  uint64
	count uint64
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
	return slice.LoadSlice(32)
}

func packageValueParse(slice *cell.Slice) ([][]byte, error) {
	return writeCellsToBuffer(slice.MustLoadRef())
}

func packageDictionaryValueParse(slice *cell.Slice) (map[int][][]byte, error) {
	slice.MustLoadUInt(256)

	cellDict := slice.MustLoadDict(16)
	dict, _ := cellDict.LoadAll()

	res := make(map[int][][]byte)

	for i, kv := range dict {
		key, _ := kv.Key.LoadSlice(32)
		value, _ := packageValueParse(kv.Value)
		log.Printf("i = %i key = %v value = %v", i, key, value)
		res[i] = value
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
