package teleportcontract

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

// serialize transaction func repeats the logic from ts code (https://github.com/RSquad/ton-teleport-btc/blob/d8cc9f0f845fc996fc2a9cf6756a504c4432ee54/utils/serialize.ts#L8)
func (c *TeleportContract) serializeTransaction(txHex string) (*cell.Cell, error) {
	offset := 0
	flags := uint8(0)
	txbuilder := cell.BeginCell()
	txbuffer, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	versionBuf := txbuffer[offset : offset+4]
	offset += 4
	txbuilder.MustStoreSlice(versionBuf, 32)

	txInBuf := readTxIn(txbuffer, &offset)

	if len(txInBuf) == 1 && txInBuf[0] == 0 {
		flags = txbuffer[offset]
		if flags != 0 {
			offset += 1
			// Store dummy + flags: 0001
			txbuilder.MustStoreSlice(txbuffer[offset-2:offset], 16)
			txInBuf = readTxIn(txbuffer, &offset)
		}
	}

	txOutBuf := readTxOut(txbuffer, &offset)

	txbuilder.MustStoreRef(splitBufferToCells(txInBuf))
	txbuilder.MustStoreRef(splitBufferToCells(txOutBuf))

	if flags&1 != 0 {
		flags ^= 1
		witnessStart := offset
		inCount, _ := readCompactSize(txInBuf, 0)

		for txin := uint64(0); txin < inCount; txin++ {
			witnessCount, size := readCompactSize(txbuffer, offset)
			offset += size

			for i := uint64(0); i < witnessCount; i++ {
				witnessLen, size := readCompactSize(txbuffer, offset)
				offset += size
				offset += int(witnessLen)
			}
		}
		witnessEnd := offset
		witnessBuf := txbuffer[witnessStart:witnessEnd]
		txbuilder.MustStoreRef(splitBufferToCells(witnessBuf))
	}

	locktimeBuf := txbuffer[offset : offset+4]
	offset += 4
	txbuilder.MustStoreSlice(locktimeBuf, 32)

	if len(txbuffer) != offset {
		return nil, fmt.Errorf("failed to parse transaction: %w", err)
	}

	return txbuilder.EndCell(), nil
}

func readTxIn(txbuffer []byte, offset *int) []byte {
	start := *offset
	inCount, size := readCompactSize(txbuffer, *offset)
	*offset += size

	for i := uint64(0); i < inCount; i++ {
		*offset += 36 // prev_hash (32) + prev_index (4)
		scriptLen, size := readCompactSize(txbuffer, *offset)
		*offset += size
		*offset += int(scriptLen) // script
		*offset += 4              // sequence
	}

	return txbuffer[start:*offset]
}

func readTxOut(txbuffer []byte, offset *int) []byte {
	start := *offset
	outCount, size := readCompactSize(txbuffer, *offset)
	*offset += size

	for i := uint64(0); i < outCount; i++ {
		*offset += 8 // value (8 bytes)
		scriptLen, size := readCompactSize(txbuffer, *offset)
		*offset += size
		*offset += int(scriptLen) // script
	}

	return txbuffer[start:*offset]
}

func readCompactSize(buffer []byte, offset int) (uint64, int) {
	size := 1
	v := buffer[offset]

	switch v {
	case 0xfd:
		return uint64(binary.LittleEndian.Uint16(buffer[offset+1 : offset+3])), size + 2
	case 0xfe:
		return uint64(binary.LittleEndian.Uint32(buffer[offset+1 : offset+5])), size + 4
	case 0xff:
		return binary.LittleEndian.Uint64(buffer[offset+1 : offset+9]), size + 8
	default:
		return uint64(v), size
	}
}

func splitBufferToCells(buffer []byte, splitSize ...int) *cell.Cell {
	sz := 1 // default splitSize
	if len(splitSize) > 0 && splitSize[0] > 0 {
		sz = splitSize[0]
	}
	if sz > 127 {
		sz = 127
	}

	cellCapacity := (127 / sz) * sz
	if cellCapacity == 0 {
		cellCapacity = 127
	}

	cellsCount := (len(buffer) + cellCapacity - 1) / cellCapacity

	if cellsCount <= 1 {
		return cell.BeginCell().MustStoreSlice(buffer, uint(len(buffer)*8)).EndCell()
	}

	cells := make([]*cell.Builder, cellsCount)
	offset := 0

	// Create cells for all chunks except last
	for i := 0; i < cellsCount-1; i++ {
		chunk := buffer[offset : offset+cellCapacity]
		cells[i] = cell.BeginCell().MustStoreSlice(chunk, uint(len(chunk)*8))
		offset += cellCapacity
	}

	// Last cell with remaining data
	remaining := buffer[offset:]
	cells[cellsCount-1] = cell.BeginCell().MustStoreSlice(remaining, uint(len(remaining)*8))

	// Link cells in reverse order (last to first)
	for i := cellsCount - 1; i >= 1; i-- {
		cells[i-1].MustStoreRef(cells[i].EndCell())
	}

	return cells[0].EndCell()
}
