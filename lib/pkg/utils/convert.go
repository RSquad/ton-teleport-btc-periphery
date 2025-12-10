package utils

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

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

func BigToUint8(x *big.Int) (uint8, error) {
	if x.Sign() < 0 {
		return 0, fmt.Errorf("value %v is negative", x)
	}

	if x.BitLen() > 8 {
		return 0, fmt.Errorf("value %v overflows uint8", x)
	}

	return uint8(x.Uint64()), nil
}

func BigToUint64(x *big.Int) (uint64, error) {
	if x.Sign() < 0 {
		return 0, fmt.Errorf("value %v is negative", x)
	}

	if x.BitLen() > 64 {
		return 0, fmt.Errorf("value %v overflows uint8", x)
	}

	return x.Uint64(), nil
}
