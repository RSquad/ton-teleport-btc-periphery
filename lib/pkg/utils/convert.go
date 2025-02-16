package utils

import (
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
