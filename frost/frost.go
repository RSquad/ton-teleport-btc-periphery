package frost

/*
#cgo CFLAGS: -I${SRCDIR}/rust
#cgo LDFLAGS: ${SRCDIR}/rust/target/release/libfrost.a
#include "rust/frost.h"
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"
)

type Package struct {
	buf []byte
}

type Pkg struct {
	identifier [32]byte
	buf        *C.uint8_t
	len        C.int
}

func NewPackage(buf []byte) Package {
	return Package{buf: buf}
}

func (p Package) ToBytes() []byte {
	return p.buf
}

func GetIdentifier(key uint16) Identifier {
	buf := make([]byte, 32)
	_ = C.ext_get_identifier(
		C.uint16_t(key),
		(*[32]C.uint8_t)(unsafe.Pointer(&buf[0])),
	)
	return Identifier(buf)
}

func DkgPart1(identifier Identifier, min_signers uint16, max_signers uint16) ([]byte, uintptr, error) {
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
		return nil, 0, Error(int32(pkgLen))
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
) (map[Identifier]Package, uintptr, []byte, error) {
	pkgs, pin := makeCPackageSlice(round1Packages)
	defer pin.Unpin()

	if len(pkgs) == 0 {
		return nil, 0, nil, errors.New("pkgs is empty")
	}

	secret := unsafe.Pointer(nil)
	r2Packages := unsafe.Pointer(nil)
	r2CulpritIdx := make([]byte, 32)
	r2PackagesLen := C.dkg_part2(
		unsafe.Pointer(r1Secret),
		(*C.Pkg)(&pkgs[0]),
		C.size_t(len(pkgs)),
		(**C.Pkg)(unsafe.Pointer(&r2Packages)),
		&secret,
		(*[32]C.uint8_t)(unsafe.Pointer(&r2CulpritIdx[0])),
	)

	round2Packages := make(map[Identifier]Package)

	if r2PackagesLen < 0 {
		return round2Packages, 0, CulpritData(r2CulpritIdx, int32(r2PackagesLen)), Error(int32(r2PackagesLen))
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
	return round2Packages, uintptr(secret), nil, nil
}

func DkgPart3(
	r2Secret uintptr,
	round1Packages map[Identifier]Package,
	round2Packages map[Identifier]Package,
) ([]byte, []byte, []byte, error) {
	r1Pkgs, pin1 := makeCPackageSlice(round1Packages)
	r2Pkgs, pin2 := makeCPackageSlice(round2Packages)
	defer pin1.Unpin()
	defer pin2.Unpin()

	if len(r1Pkgs) == 0 {
		return nil, nil, nil, errors.New("r1Pkgs is empty")
	}

	if len(r2Pkgs) == 0 {
		return nil, nil, nil, errors.New("r2Pkgs is empty")
	}

	secretPkgPtr := unsafe.Pointer(nil)
	secretPkgLen := C.size_t(0)
	publicPkgPtr := unsafe.Pointer(nil)
	publicPkgLen := C.size_t(0)
	r3CulpritIdx := make([]byte, 32)
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
		(*[32]C.uint8_t)(unsafe.Pointer(&r3CulpritIdx[0])),
	)

	if ret < 0 {
		return nil, nil, CulpritData(r3CulpritIdx, int32(ret)), Error(int32(ret))
	}

	publicKeyPackage := make([]byte, publicPkgLen)
	KeyPackage := make([]byte, secretPkgLen)
	copy(publicKeyPackage, unsafe.Slice((*byte)(publicPkgPtr), publicPkgLen))
	copy(KeyPackage, unsafe.Slice((*byte)(secretPkgPtr), secretPkgLen))
	C.free_package_ptr((*C.uint8_t)(publicPkgPtr), publicPkgLen)
	C.free_package_ptr((*C.uint8_t)(secretPkgPtr), secretPkgLen)

	return KeyPackage, publicKeyPackage, nil, nil
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
	keyPackage Package,
	message []byte,
	commitments map[Identifier]Package,
	nonces Package,
	merkleRoot []byte,
) ([]byte, []byte, error) {
	commitmentsPkgs, commitmentsPin := makeCPackageSlice(commitments)
	defer commitmentsPin.Unpin()

	signingShares := newEmptyBuffer()
	culpritIdx := make([]byte, 32)

	ret := C.sign_with_tweak(
		newBufferFromPackage(keyPackage),
		newBufferFromSlice(message),
		(*C.Pkg)(&commitmentsPkgs[0]),
		C.size_t(len(commitmentsPkgs)),
		newBufferFromPackage(nonces),
		newBufferFromSlice(merkleRoot),
		&signingShares,
		(*[32]C.uint8_t)(unsafe.Pointer(&culpritIdx[0])),
	)

	if ret < 0 {
		return nil, CulpritData(culpritIdx, int32(ret)), Error(int32(ret))
	}

	return extractSlice(signingShares), nil, nil
}

func AggregateWithTweak(
	message []byte,
	commitments map[Identifier]Package,
	signatureShares map[Identifier]Package,
	pubkeyPackage Package,
	merkleRoot []byte,
) ([]byte, []byte, error) {
	commitmentsPkgs, commitmentsPin := makeCPackageSlice(commitments)
	signatureSharesPkgs, signatureSharesPin := makeCPackageSlice(signatureShares)

	defer commitmentsPin.Unpin()
	defer signatureSharesPin.Unpin()

	signature := newEmptyBuffer()
	culpritIdx := make([]byte, 32)

	ret := C.aggregate_with_tweak(
		newBufferFromSlice(message),
		(*C.Pkg)(&commitmentsPkgs[0]),
		C.size_t(len(commitmentsPkgs)),
		(*C.Pkg)(&signatureSharesPkgs[0]),
		C.size_t(len(signatureSharesPkgs)),
		newBufferFromPackage(pubkeyPackage),
		newBufferFromSlice(merkleRoot),
		&signature,
		(*[32]C.uint8_t)(unsafe.Pointer(&culpritIdx[0])),
	)

	if ret < 0 {
		return nil, CulpritData(culpritIdx, int32(ret)), Error(int32(ret))
	}

	return extractSlice(signature), nil, nil
}

func Verify(
	pubkeyPackage Package,
	message []byte,
	signature []byte,
	merkleRoot []byte,
) (bool, error) {
	ret := C.verify(
		newBufferFromSlice(message),
		newBufferFromPackage(pubkeyPackage),
		newBufferFromSlice(signature),
		newBufferFromSlice(merkleRoot),
	)

	if ret < 0 {
		return false, Error(int32(ret))
	}

	return true, nil
}

func ExtractPublicKeyFromPackage(pkg []byte) ([]byte, error) {
	publicKey := newEmptyBuffer()
	ret := C.extract_public_key_from_package(
		newBufferFromSlice(pkg),
		&publicKey,
	)
	if ret < 0 {
		return nil, Error(int32(ret))
	}
	return extractSlice(publicKey), nil
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

func FreeR1Secret(secret uintptr) {
	C.free_r1_secret(unsafe.Pointer(secret))
}

func FreeR2Secret(secret uintptr) {
	C.free_r2_secret(unsafe.Pointer(secret))
}
