package coordinator

import (
	"errors"
	"reflect"
	"testing"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestLoadSharesMap(t *testing.T) {
	idxToCell := func(idx uint16) *cell.Cell {
		return cell.BeginCell().MustStoreUInt(uint64(idx), 64).EndCell()
	}
	bufToCell := func(buf []byte) *cell.Cell {
		return cell.BeginCell().MustStoreSlice(buf, uint(len(buf)*8)).EndCell()
	}
	dict := cell.NewDict(64)
	dict.Set(idxToCell(1), bufToCell([]byte{1, 2, 3}))
	dict.Set(idxToCell(2), bufToCell([]byte{4, 5, 6}))
	dict.Set(idxToCell(3), bufToCell([]byte{7, 8, 9}))

	dictCell := dict.AsCell()
	if dictCell == nil {
		dictCell = cell.BeginCell().EndCell()
	}
	c := cell.BeginCell().MustStoreRef(dictCell).EndCell()
	testCases := []struct {
		name      string
		inputCell *cell.Cell
		want      map[uint16][]byte
		err       error
	}{
		{
			name:      "test1",
			inputCell: c,
			want: map[uint16][]byte{
				1: {1, 2, 3},
				2: {4, 5, 6},
				3: {7, 8, 9},
			},
		},
		{
			name:      "test2",
			inputCell: cell.BeginCell().MustStoreRef(cell.BeginCell().MustStoreDict(dict).EndCell()).EndCell(),
			want:      map[uint16][]byte{},
			err:       errors.New("failed to load dict"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadSharesMap(tc.inputCell.BeginParse())
			if tc.err != nil {
				if err == nil {
					t.Fatalf("error expected = %v, got %v", tc.err, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("loadSharesMap(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
