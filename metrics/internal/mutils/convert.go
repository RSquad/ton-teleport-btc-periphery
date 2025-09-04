package mutils

import "github.com/btcsuite/btcd/chaincfg/chainhash"

func BytesToBTCHash(b []byte) *chainhash.Hash {
	if len(b) > chainhash.HashSize {
		b = b[:chainhash.HashSize]
	}

	var h chainhash.Hash

	copy(h[:], b)
	return &h
}
