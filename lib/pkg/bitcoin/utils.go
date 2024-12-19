package bitcoin

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/xssnick/tonutils-go/address"
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

func TxToHex(tx *wire.MsgTx) (string, error) {
	var buf bytes.Buffer
	err := tx.Serialize(&buf)
	if err != nil {
		return "", fmt.Errorf("failed to serialize transaction: %w", err)
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func HexToTx(txHex string, version int32) (*wire.MsgTx, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return nil, err
	}

	tx := wire.NewMsgTx(version)

	err = tx.Deserialize(bytes.NewReader(txBytes))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func TonAddrToBytesForTapLeaf(addr *address.Address) []byte {
	suffix := byte(0x00)
	if addr.Workchain() != 0 {
		suffix = byte(0xff)
	}

	return append(addr.Data(), suffix)
}

func TxContainsOutWithAddr(tx *btcjson.TxRawResult, addr string) (bool, *btcjson.Vout) {
	for _, vout := range tx.Vout {
		if vout.ScriptPubKey.Address == addr {
			return true, &vout
		}
	}
	return false, nil
}
