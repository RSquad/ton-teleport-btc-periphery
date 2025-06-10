package coordinator

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/xssnick/tonutils-go/tvm/cell"
)

func TestAttachSessionSignatureToCell(t *testing.T) {
	tests := []struct {
		name      string
		body      *cell.Cell
		signature []byte
		wantError bool
	}{
		{
			name:      "Valid signature with simple body",
			body:      createTestBody(1, time.Now().Unix(), make([]byte, 32), make([]byte, 32)),
			signature: make64ByteSignature(),
			wantError: false,
		},
		{
			name:      "Valid signature with complex body",
			body:      createTestBody(1, time.Now().Unix(), make([]byte, 32), make([]byte, 32)),
			signature: make64ByteSignature(),
			wantError: false,
		},
		{
			name:      "Empty signature",
			body:      createTestBody(1, time.Now().Unix(), make([]byte, 32), make([]byte, 32)),
			signature: []byte{},
			wantError: true,
		},
		{
			name:      "Nil body should be handled gracefully",
			body:      nil,
			signature: make64ByteSignature(),
			wantError: true, // Should return error for nil body
		},
		{
			name:      "Invalid signature length (not 64 bytes)",
			body:      createTestBody(1, time.Now().Unix(), make([]byte, 32), make([]byte, 32)),
			signature: make([]byte, 32), // Wrong length
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test normal execution - should not panic
			result, err := attachSessionSignatureToCell(tt.body, tt.signature)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if result != nil {
					t.Errorf("Expected nil result when error occurs")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result == nil {
				t.Fatalf("Result should not be nil")
			}

			// Verify the structure of the resulting cell
			verifyAttachedSignatureCell(t, result, tt.body, tt.signature)
		})
	}
}

func TestAttachSessionSignatureToCell_RealSignature(t *testing.T) {
	// Generate a real Ed25519 key pair for more realistic testing
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Ed25519 key pair: %v", err)
	}
	body := createTestBody(1, time.Now().Unix(), publicKey[:], make([]byte, 32))
	signature := body.Sign(privateKey)
	if len(signature) != 64 {
		t.Fatalf("Ed25519 signature should be 64 bytes, got %d", len(signature))
	}

	result, err := attachSessionSignatureToCell(body, signature)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Fatalf("Result should not be nil")
	}

	// Verify the structure
	verifyAttachedSignatureCell(t, result, body, signature)

	bs := result.BeginParse()
	realSignature := bs.MustLoadRef().MustLoadSlice(512)
	bs.MustLoadSlice(32 + 32 + 16 + 32)
	pubkey := bs.MustLoadSlice(256)
	if !bytes.Equal(pubkey, publicKey[:]) {
		t.Fatalf("Public key should match")
	}
	if !ed25519.Verify(publicKey, body.Hash(), realSignature) {
		t.Fatalf("Failed to verify signature")
	}
}

func TestAttachSessionSignatureToCell_Structure(t *testing.T) {
	body := createTestBody(1, time.Now().Unix(), make([]byte, 32), make([]byte, 32))
	signature := make64ByteSignature()

	result, err := attachSessionSignatureToCell(body, signature)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	slice := result.BeginParse()

	if slice.RefsNum() < 1 {
		t.Fatalf("Should have at least one reference, got %d", slice.RefsNum())
	}
	signatureCell := slice.MustLoadRef()

	// Verify signature cell contains the signature
	extractedSig := signatureCell.MustLoadSlice(512) // 64 bytes * 8 bits
	if !bytes.Equal(signature, extractedSig) {
		t.Errorf("Extracted signature should match original")
	}

	if slice.BitsLeft() == 0 && slice.RefsNum() == 0 {
		t.Errorf("Should have remaining body content")
	}
}

// Helper functions

func createTestBody(validatorIdx uint16, dkgUntil int64, sessionPupkey []byte, pubkeyPackage []byte) *cell.Cell {
	return BuildSendRound3Body(60, validatorIdx, dkgUntil, sessionPupkey, pubkeyPackage)
}

func make64ByteSignature() []byte {
	signature := make([]byte, 64)
	for i := range signature {
		signature[i] = byte(i % 256)
	}
	return signature
}

func verifyAttachedSignatureCell(t *testing.T, result *cell.Cell, originalBody *cell.Cell, signature []byte) {
	slice := result.BeginParse()
	if slice.RefsNum() < 1 {
		t.Fatalf("Should have at least one reference for signature, got %d", slice.RefsNum())
	}
	signatureCell := slice.MustLoadRef()

	if len(signature) > 0 {
		extractedSig := signatureCell.MustLoadSlice(uint(len(signature) * 8))
		if !bytes.Equal(signature, extractedSig) {
			t.Errorf("Extracted signature should match original")
		}
	}
	hasRemainingContent := slice.BitsLeft() > 0 || slice.RefsNum() > 0
	if !hasRemainingContent {
		t.Errorf("Should have remaining body content")
	}
}
