package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/xssnick/tonutils-go/address"
)

func AddrToRawString(addr *address.Address) string {
	return fmt.Sprintf("%d:%s", addr.Workchain(), hex.EncodeToString(addr.Data()))
}

func MustHexToBytes(s string, size int) []byte {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return BytesPadTo(bytes, size)
}

func MustBase64ToHexStr(s string, size int) string {
	bytes, _ := base64.StdEncoding.DecodeString(s)
	if size > 0 {
		return fmt.Sprintf("%x", BytesPadTo(bytes, size))
	}
	return fmt.Sprintf("%x", bytes)
}
