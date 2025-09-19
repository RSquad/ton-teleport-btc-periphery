package mutils

import (
	"fmt"
	"strings"

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

func ExtractMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func JoinToStr[K ~string](keys []K) string {
	if len(keys) == 0 {
		return ""
	}

	strs := make([]string, len(keys))
	for i, k := range keys {
		strs[i] = string(k)
	}
	return strings.Join(strs, ",")
}
