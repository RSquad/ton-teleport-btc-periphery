package bitcoin

import (
    "bytes"

    "github.com/btcsuite/btcd/chaincfg/chainhash"
    "github.com/btcsuite/btcd/wire"
)

func SliceOfHashesContains(haystack []*chainhash.Hash, needle *chainhash.Hash) bool {
    for _, item := range haystack {
        if item.IsEqual(needle) {
            return true
        }
    }
    return false
}

func BlockHeaderToBytes(header *wire.BlockHeader) ([]byte, error) {
    var buf bytes.Buffer
    err := header.Serialize(&buf)
    return buf.Bytes(), err
}
