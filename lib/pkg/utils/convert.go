package utils

import "encoding/hex"

func BytesToHexString(b []byte) string {
    return hex.EncodeToString(b)
}
