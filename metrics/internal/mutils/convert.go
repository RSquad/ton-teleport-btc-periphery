package mutils

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func BytesToBTCHash(b []byte) *chainhash.Hash {
	if len(b) > chainhash.HashSize {
		b = b[:chainhash.HashSize]
	}

	var h chainhash.Hash

	copy(h[:], b)
	return &h
}

func NanoIntToString(n int64) string {
	whole := n / 1_000_000_000
	frac := n % 1_000_000_000
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%09d", whole, frac)
}
