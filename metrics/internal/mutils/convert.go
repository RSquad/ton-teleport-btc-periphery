package mutils

import (
	"bytes"
	"fmt"
	"html"
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

func CreateShortLink(text, url string) string {
	escapedText := html.EscapeString(text)
	escapedURL := html.EscapeString(url)

	return fmt.Sprintf("<a href=\"%s\">%s</a>", escapedURL, escapedText)
}

func BtcExplorerLink(address string) string {
	cfg := config.GetGlobalRuntimeConfig()
	explorerUrl := "http://btc"

	if cfg != nil {
		explorerUrl = cfg.BtcExplorer()
	}

	return CreateShortLink("link",
		fmt.Sprintf(
			"%s/%s",
			explorerUrl,
			address,
		),
	)
}

func TonExplorerLink(address string) string {
	cfg := config.GetGlobalRuntimeConfig()
	explorerUrl := "http://ton"

	if cfg != nil {
		explorerUrl = cfg.TonExplorer()
	}

	return CreateShortLink("link",
		fmt.Sprintf(
			"%s/%s",
			explorerUrl,
			address,
		),
	)
}

func RunbookLink(alertName string) string {
	cfg := config.GetGlobalRuntimeConfig()
	runbookUrl := "http://runbook"

	if cfg != nil {
		runbookUrl = cfg.RunbookUrl()
	}

	return CreateShortLink("link",
		fmt.Sprintf(
			"%s/%s.md",
			runbookUrl,
			alertName,
		),
	)
}
