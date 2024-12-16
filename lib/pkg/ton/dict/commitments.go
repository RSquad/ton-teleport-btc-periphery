package dict

// import (
// 	"fmt"

// 	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
// 	"github.com/xssnick/tonutils-go/tvm/cell"
// )

// type Commitments map[string][][]byte

// func parseCommitmentKey(keySlice *cell.Slice, keySize uint) string {
// 	key := keySlice.MustLoadSlice(keySize)

// 	return fmt.Sprintf("%x", key)
// }

// func parseCommitmentValue(value *cell.Slice) ([]byte, error) {
// 	return utils.WriteSlicesToBuffer(value.MustLoadRef()), nil
// }

// func NewCommitments(dict *cell.Dictionary) (*Commitments, error) {
// 	result, err := parseddict.New(dict, parseCommitmentKey, parseCommitmentValue)
// 	if err != nil {
// 		return nil, err
// 	}
// 	commitments := Commitments(*result)
// 	return &commitments, nil
// }
