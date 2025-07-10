package dkg

import (
	"crypto/rand"
	"testing"
)

func verifyDKGExecutionArtifactsIsClear(artifacts *ExecutionArtifacts, t *testing.T) {
	if artifacts.r1 != nil {
		t.Error("artifacts.r1 is expected to be nil")
		return
	}

	if artifacts.r2 != nil {
		t.Error("artifacts.r2 is expected to be nil")
		return
	}

	if artifacts.r3 != nil {
		t.Error("artifacts.r3 is expected to be nil")
		return
	}
}

func isAllZero(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

func TestDKGExecutionArtifactsCleanupR1(t *testing.T) {
	artifacts := ExecutionArtifacts{}

	artifacts.Cleanup()
	verifyDKGExecutionArtifactsIsClear(&artifacts, t)

	pkgTmpBuf := make([]byte, 32)
	rand.Read(pkgTmpBuf)

	var r2PublicX25519TmpBuf [32]byte
	rand.Read(r2PublicX25519TmpBuf[:])

	var r2PrivateX25519TmpBuf [32]byte
	rand.Read(r2PrivateX25519TmpBuf[:])

	artifacts.r1 = &Round1Result{
		pkg:             pkgTmpBuf,
		secret:          NewSecret(0),
		r2PublicX25519:  &r2PublicX25519TmpBuf,
		r2PrivateX25519: &r2PrivateX25519TmpBuf,
	}

	if artifacts.r1 == nil {
		t.Error("artifacts.r1 is nil")
		return
	}

	if len(artifacts.r1.r2PrivateX25519) != 32 {
		t.Error("Length of artifacts.r1 is expected to be 32")
		return
	}

	if isAllZero(artifacts.r1.r2PrivateX25519[:]) {
		t.Error("artifacts.r1 is filled with zeros")
		return
	}

	artifacts.SafeCleanPrivateX25519()

	if !isAllZero(artifacts.r1.r2PrivateX25519[:]) {
		t.Error("artifacts.r1 is expected to be filled with zeros")
		return
	}
}
