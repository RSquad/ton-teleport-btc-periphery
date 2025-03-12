package frost

/*
#include "rust/frost.h"
#include <stdlib.h>
*/
import "C"

import (
	"unsafe"
)

func newBuffer(p unsafe.Pointer, len int) C.Buffer {
	return C.Buffer{(*C.uint8_t)(p), C.size_t(len)}
}

func newEmptyBuffer() C.Buffer {
	return newBuffer(unsafe.Pointer(nil), 0)
}

func newBufferFromSlice(s []byte) C.Buffer {
	return newBuffer(unsafe.Pointer(&s[0]), len(s))
}

func newBufferFromPackage(p Package) C.Buffer {
	return newBufferFromSlice(p.buf)
}

func extractSlice(buf C.Buffer) []byte {
	s := make([]byte, buf.len)
	copy(s, unsafe.Slice((*byte)(buf.data), buf.len))
	C.free_package_ptr((*C.uint8_t)(buf.data), buf.len)
	return s
}
