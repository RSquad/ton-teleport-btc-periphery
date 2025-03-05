package frost

/*
#cgo CFLAGS: -I./rust
#cgo LDFLAGS: ./rust/target/debug/libfrost.a
#include "./rust/frost.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"runtime"
	"unsafe"
)

type Package struct {
	buf []byte
}

type Identifier [32]byte

type Pkg struct {
	identifier [32]byte
	buf        *C.uint8_t
	len        C.int
}

func NewPackage(buf []byte) Package {
	return Package{buf: buf}
}

func IdentifierToBytes(ident Identifier) []byte {
	return ident[:]
}

func newBuffer(p unsafe.Pointer, len int) C.Buffer {
	return C.Buffer{(*C.uint8_t)(p), C.size_t(len)}
}

func PackageToBytes(p Package) []byte {
	return p.buf
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

func GetIdentifierFromHex(hexStr string) []byte {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		panic(err)
	}
	return data
}

func DkgPart1(identifier []byte, min_signers uint16, max_signers uint16) ([]byte, uintptr, error) {
	var secret unsafe.Pointer
	pkgLen := C.int(0)
	pkgLen = C.dkg_part1(
		(*[32]C.uint8_t)(unsafe.Pointer(&identifier[0])),
		C.uint16_t(min_signers),
		C.uint16_t(max_signers),
		(*C.uint8_t)(unsafe.Pointer(nil)),
		pkgLen,
		&secret,
	)
	if pkgLen < 0 {
		return nil, 0, fmt.Errorf("dkg_part1 error %d", pkgLen)
	}
	pkg := make([]byte, pkgLen)

	C.dkg_part1(
		(*[32]C.uint8_t)(unsafe.Pointer(&identifier[0])),
		C.uint16_t(min_signers),
		C.uint16_t(max_signers),
		(*C.uint8_t)(unsafe.Pointer(&pkg[0])),
		pkgLen,
		&secret,
	)

	return pkg, uintptr(secret), nil
}

func DkgPart2(
	r1Secret uintptr,
	round1Packages map[Identifier]Package,
) (map[Identifier]Package, uintptr, error) {
	pkgs, pin := makeCPackageSlice(round1Packages)
	secret := unsafe.Pointer(nil)
	r2Packages := unsafe.Pointer(nil)
	r2PackagesLen := C.dkg_part2(
		unsafe.Pointer(r1Secret),
		(*C.Pkg)(&pkgs[0]),
		C.size_t(len(pkgs)),
		(**C.Pkg)(unsafe.Pointer(&r2Packages)),
		&secret,
	)
	pin.Unpin()
	round2Packages := make(map[Identifier]Package)

	if r2PackagesLen <= 0 {
		return round2Packages, 0, fmt.Errorf("dkg_part2 error %d", r2PackagesLen)
	}
	pkgs = unsafe.Slice((*C.Pkg)(r2Packages), r2PackagesLen)
	for _, v := range pkgs {
		id := Identifier{}
		copy(id[:], unsafe.Slice((*byte)(unsafe.Pointer(&v.identifier[0])), 32))
		pkg := make([]byte, v.len)
		copy(pkg, unsafe.Slice((*byte)(unsafe.Pointer(v.buf)), v.len))
		round2Packages[id] = Package{pkg}
	}
	C.free_r2_pkg_vec((*C.Pkg)(r2Packages), C.size_t(r2PackagesLen))
	return round2Packages, uintptr(secret), nil
}

func DkgPart3(
	r2Secret uintptr,
	round1Packages map[Identifier]Package,
	round2Packages map[Identifier]Package,
) ([]byte, []byte, error) {
	r1Pkgs, pin1 := makeCPackageSlice(round1Packages)
	r2Pkgs, pin2 := makeCPackageSlice(round2Packages)
	secretPkgPtr := unsafe.Pointer(nil)
	secretPkgLen := C.size_t(0)
	publicPkgPtr := unsafe.Pointer(nil)
	publicPkgLen := C.size_t(0)
	ret := C.dkg_part3(
		unsafe.Pointer(r2Secret),
		(*C.Pkg)(&r1Pkgs[0]),
		C.size_t(len(r1Pkgs)),
		(*C.Pkg)(&r2Pkgs[0]),
		C.size_t(len(r2Pkgs)),
		(**C.uint8_t)(unsafe.Pointer(&publicPkgPtr)),
		&publicPkgLen,
		(**C.uint8_t)(unsafe.Pointer(&secretPkgPtr)),
		&secretPkgLen,
	)

	pin1.Unpin()
	pin2.Unpin()

	if ret < 0 {
		return nil, nil, fmt.Errorf("%d", ret)
	}

	publicKeyPackage := make([]byte, publicPkgLen)
	KeyPackage := make([]byte, secretPkgLen)
	copy(publicKeyPackage, unsafe.Slice((*byte)(publicPkgPtr), publicPkgLen))
	copy(KeyPackage, unsafe.Slice((*byte)(secretPkgPtr), secretPkgLen))
	C.free_package_ptr((*C.uint8_t)(publicPkgPtr), publicPkgLen)
	C.free_package_ptr((*C.uint8_t)(secretPkgPtr), secretPkgLen)

	return KeyPackage, publicKeyPackage, nil
}

func makeCPackageSlice(packages map[Identifier]Package) ([]C.Pkg, runtime.Pinner) {
	pinner := runtime.Pinner{}
	i := 0
	pkgs := make([]C.Pkg, len(packages))
	for sender, pkg := range packages {
		bufPtr := (*C.uint8_t)(unsafe.Pointer(&pkg.buf[0]))
		bufLen := C.size_t(len(pkg.buf))
		// If we will not pin the pointer then
		// we'll get runtime error
		// `cgo argument has Go pointer to unpinned Go pointer`
		pinner.Pin(bufPtr)
		pkgs[i] = C.Pkg{
			*(*[32]C.uint8_t)(unsafe.Pointer(&sender[0])),
			bufPtr,
			bufLen,
		}
		i += 1
	}
	return pkgs, pinner
}

func Commit(keyPackage Package) ([]byte, []byte, error) {
	nonces := newEmptyBuffer()
	commitments := newEmptyBuffer()
	ret := C.commit(
		newBufferFromPackage(keyPackage),
		&nonces,
		&commitments,
	)

	if ret < 0 {
		return nil, nil, fmt.Errorf("%d", ret)
	}

	return extractSlice(nonces), extractSlice(commitments), nil
}

func SignWithTweak(
	// merkleRoot []byte,
	keyPackage Package,
	message []byte,
	commitments map[Identifier]Package,
	nonces Package,
) ([]byte, error) {
	commitmentsPkgs, commitmentsPin := makeCPackageSlice(commitments)
	signatureShares := newEmptyBuffer()
	ret := C.sign_with_tweak(
		// (*[32]C.uint8_t)(unsafe.Pointer(&merkleRoot[0])),
		newBufferFromPackage(keyPackage),
		newBufferFromSlice(message),
		(*C.Pkg)(&commitmentsPkgs[0]),
		C.size_t(len(commitmentsPkgs)),
		newBufferFromPackage(nonces),
		&signatureShares,
	)
	commitmentsPin.Unpin()

	if ret < 0 {
		return nil, fmt.Errorf("%d", ret)
	}

	return extractSlice(signatureShares), nil
}

func AggregateWithTweak(
	// merkleRoot []byte,
	message []byte,
	commitments map[Identifier]Package,
	signatureShares map[Identifier]Package,
	pubkeyPackage Package,
) ([]byte, error) {
	commitmentsPkgs, commitmentsPin := makeCPackageSlice(commitments)
	signatureSharesPkgs, signatureSharesPin := makeCPackageSlice(signatureShares)
	signature := newEmptyBuffer()

	ret := C.aggregate_with_tweak(
		// (*[32]C.uint8_t)(unsafe.Pointer(&merkleRoot[0])),
		newBufferFromSlice(message),
		(*C.Pkg)(&commitmentsPkgs[0]),
		C.size_t(len(commitmentsPkgs)),
		(*C.Pkg)(&signatureSharesPkgs[0]),
		C.size_t(len(signatureSharesPkgs)),
		newBufferFromPackage(pubkeyPackage),
		&signature,
	)

	commitmentsPin.Unpin()
	signatureSharesPin.Unpin()

	if ret < 0 {
		return nil, fmt.Errorf("%d", ret)
	}

	return extractSlice(signature), nil
}

func Verify(
	pubkeyPackage Package,
	message []byte,
	signature []byte,
	// merkleRoot []byte,
) (bool, error) {
	ret := C.verify(
		// (*[32]C.uint8_t)(unsafe.Pointer(&merkleRoot[0])),
		newBufferFromSlice(message),
		newBufferFromPackage(pubkeyPackage),
		newBufferFromSlice(signature),
	)

	if ret < 0 {
		return false, fmt.Errorf("%d", ret)
	}

	return true, nil
}
