package frost

/*
#cgo CFLAGS: -I${SRCDIR}/rust
#cgo LDFLAGS: ${SRCDIR}/rust/target/debug/libfrost.a
#include "rust/frost.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/hex"
	"unsafe"
)

type Identifier [32]byte

func DecodeIdentifier(s string) (*Identifier, error) {
	var identifier Identifier
	id, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	copy(identifier[:], id)
	return &identifier, nil
}

func GetIdentifier(key uint16) []byte {
	buf := make([]byte, 32)
	_ = C.ext_get_identifier(
		C.uint16_t(key),
		(*[32]C.uint8_t)(unsafe.Pointer(&buf[0])),
	)
	return buf
}

func (id Identifier) ToBytes() []byte {
	return id[:]
}
