package mutils

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
)

func BytesToBTCHash(b []byte) *chainhash.Hash {
	bb := bytes.Clone(b)

	if len(bb) > chainhash.HashSize {
		bb = bb[:chainhash.HashSize]
	}

	var h chainhash.Hash

	slices.Reverse(bb)
	copy(h[:], bb)
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

func ExtractMapKeysConv[K comparable, Val any, Out any](m map[K]Val, conv func(K) Out) []Out {
	out := make([]Out, 0, len(m))
	for k := range m {
		out = append(out, conv(k))
	}
	return out
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

func BtcExplorerLink(address string) string {
	cfg := config.GetGlobalRuntimeConfig()
	explorerUrl := "http://btc"

	if cfg != nil {
		explorerUrl = cfg.BtcExplorer()
	}

	return fmt.Sprintf(
		"<a href=\"%s/%s\">%s</a>",
		explorerUrl,
		address,
		address,
	)
}

func TonExplorerLink(address string) string {
	cfg := config.GetGlobalRuntimeConfig()
	explorerUrl := "http://ton"

	if cfg != nil {
		explorerUrl = cfg.TonExplorer()
	}

	return fmt.Sprintf(
		"<a href=\"%s/%s\">%s</a>",
		explorerUrl,
		address,
		address,
	)
}
