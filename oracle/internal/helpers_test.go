package helpers

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"testing"
)

func TestCalcMinSigners(t *testing.T) {
	tests := []struct {
		name       string
		maxSigners uint16
		want       uint16
		wantErr    bool
	}{
		{
			name:       "maxSigners = 0",
			maxSigners: 0,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "maxSigners = 1",
			maxSigners: 1,
			want:       0,
			wantErr:    true,
		},
		{
			name:       "maxSigners = 2",
			maxSigners: 2,
			want:       2,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 3",
			maxSigners: 3,
			want:       2,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 4",
			maxSigners: 4,
			want:       2,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 5",
			maxSigners: 5,
			want:       3,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 6",
			maxSigners: 6,
			want:       4,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 7",
			maxSigners: 7,
			want:       4,
			wantErr:    false,
		},
		{
			name:       "maxSigners = 9",
			maxSigners: 9,
			want:       6,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcMinSigners(tt.maxSigners)
			if (err != nil) != tt.wantErr {
				t.Errorf("CalcMinSigners(%d) error = %v, wantErr %v", tt.maxSigners, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("CalcMinSigners(%d) = %d, want %d", tt.maxSigners, got, tt.want)
			}
		})
	}
}

func TestExtractExitCode(t *testing.T) {
	tests := []struct {
		name         string
		errorLog     string
		expectedCode int
		expectError  bool
	}{
		{
			name:         "Extract exitcode from standard format",
			errorLog:     "Cannot run message on account: inbound external message rejected by transaction B0279A0FE4EF5A2759AFDCD421FF5213215357E9D932564B9B921B9B88EC6018: exitcode=114, steps=133, gas_used=0",
			expectedCode: 114,
			expectError:  false,
		},
		{
			name:         "No exitcode in error log",
			errorLog:     "Some other error without exitcode",
			expectedCode: 0,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := ExtractExitCode(tt.errorLog)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if code != tt.expectedCode {
					t.Errorf("Expected exitcode %d, got %d", tt.expectedCode, code)
				}
			}
		})
	}
}

func generateCommitment(idx int) []byte {
	commitment := make([]byte, FrostCommitmentLength)
	for i := range commitment {
		commitment[i] = byte((idx + i) % 256)
	}
	return commitment
}

func TestSerializeCommitments(t *testing.T) {
	commitment1 := generateCommitment(0)
	commitment2 := generateCommitment(1)
	commitment3 := generateCommitment(2)

	serialized3 := append([]byte{3}, commitment1...)
	serialized3 = append(serialized3, commitment2...)
	serialized3 = append(serialized3, commitment3...)

	tests := []struct {
		name        string
		commitments [][]byte
		expected    []byte
		expectError bool
	}{
		{
			name:        "Empty commitments",
			commitments: [][]byte{},
			expected:    []byte{0}, // count = 0
			expectError: false,
		},
		{
			name:        "Single commitment with correct length",
			commitments: [][]byte{commitment1},
			expected:    append([]byte{1}, commitment1...), // count=1, then commitment
			expectError: false,
		},
		{
			name: "Multiple commitments with correct length",
			commitments: [][]byte{
				commitment1,
				commitment2,
				commitment3,
			},
			expected:    serialized3, // count=3, then 3 commitments
			expectError: false,
		},
		{
			name:        "Commitment with incorrect length should return error",
			commitments: [][]byte{{0x01, 0x02, 0x03}}, // wrong length
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Zero length commitment should return error",
			commitments: [][]byte{{}}, // wrong length
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SerializeCommitments(tt.commitments, FrostCommitmentLength)

			if tt.expectError {
				if err == nil {
					t.Errorf("SerializeCommitments() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("SerializeCommitments() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("SerializeCommitments() length = %d, expected %d", len(result), len(tt.expected))
				return
			}
			for i, b := range result {
				if b != tt.expected[i] {
					t.Errorf("SerializeCommitments() byte %d = %d, expected %d", i, b, tt.expected[i])
				}
			}
		})
	}
}

func TestDeserializeCommitments(t *testing.T) {
	commitment1 := generateCommitment(0)
	commitment2 := generateCommitment(1)
	commitment3 := generateCommitment(2)

	serialized3 := append([]byte{3}, commitment1...)
	serialized3 = append(serialized3, commitment2...)
	serialized3 = append(serialized3, commitment3...)

	serialized2 := append([]byte{2}, commitment1...)
	serialized2 = append(serialized2, commitment2...)

	tests := []struct {
		name          string
		serialized    []byte
		expectedCount int
		expected      [][]byte
		expectError   bool
		errorContains string
	}{
		{
			name:          "Empty serialized data",
			serialized:    []byte{},
			expectedCount: 0,
			expected:      nil,
			expectError:   true,
			errorContains: "serialized commitments data is empty",
		},
		{
			name:          "Single commitment",
			serialized:    append([]byte{1}, commitment1...),
			expectedCount: 1,
			expected:      [][]byte{commitment1},
			expectError:   false,
		},
		{
			name:          "Multiple commitments",
			serialized:    serialized3,
			expectedCount: 3,
			expected:      [][]byte{commitment1, commitment2, commitment3},
			expectError:   false,
		},
		{
			name:          "Zero commitments",
			serialized:    []byte{0},
			expectedCount: 0,
			expected:      [][]byte{},
			expectError:   false,
		},
		{
			name:          "Incorrect count - expected more",
			serialized:    append([]byte{1}, commitment1...),
			expectedCount: 2,
			expected:      nil,
			expectError:   true,
			errorContains: "incorrect number of commitments: expected 2, got 1",
		},
		{
			name:          "Incorrect count - expected less",
			serialized:    serialized2,
			expectedCount: 1,
			expected:      nil,
			expectError:   true,
			errorContains: "incorrect number of commitments: expected 1, got 2",
		},
		{
			name:          "Insufficient data for commitments",
			serialized:    append([]byte{2}, commitment1...), // says 2 commitments but only has data for 1
			expectedCount: 2,
			expected:      nil,
			expectError:   true,
			errorContains: "insufficient data: cannot read commitments",
		},
		{
			name:          "Partial commitment data",
			serialized:    append([]byte{1}, commitment1[:len(commitment1)-1]...),
			expectedCount: 1,
			expected:      nil,
			expectError:   true,
			errorContains: "insufficient data: cannot read commitments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize test data in serialized commitments
			if !tt.expectError && len(tt.serialized) > 1 {
				count := int(tt.serialized[0])
				for i := 0; i < count; i++ {
					for j := 0; j < FrostCommitmentLength; j++ {
						tt.serialized[1+i*FrostCommitmentLength+j] = byte((i + j) % 256)
					}
				}
			}

			// Initialize expected result with the same test data
			if !tt.expectError && len(tt.expected) > 0 {
				for i, commitment := range tt.expected {
					for j := range commitment {
						commitment[j] = byte((i + j) % 256)
					}
				}
			}

			result, err := DeserializeCommitments(tt.serialized, tt.expectedCount, FrostCommitmentLength)

			if tt.expectError {
				if err == nil {
					t.Errorf("DeserializeCommitments() expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("DeserializeCommitments() error = %v, expected to contain %q", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("DeserializeCommitments() unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("DeserializeCommitments() length = %d, expected %d", len(result), len(tt.expected))
				return
			}

			for i, commitment := range result {
				if len(commitment) != FrostCommitmentLength {
					t.Errorf("DeserializeCommitments() commitment %d length = %d, expected %d", i, len(commitment), FrostCommitmentLength)
					continue
				}
				for j, b := range commitment {
					if b != tt.expected[i][j] {
						t.Errorf("DeserializeCommitments() commitment %d byte %d = %d, expected %d", i, j, b, tt.expected[i][j])
					}
				}
			}
		})

		t.Run("DeserializeInputCommitmentForAll", func(t *testing.T) {
			commitsMap := map[uint16][]byte{0: generateCommitment(0), 1: generateCommitment(1), 2: generateCommitment(2)}
			result, err := DeserializeInputCommitmentForAll(commitsMap, 3, -1)
			if err == nil {
				t.Errorf("expected error but got none")
			}
			if result != nil {
				t.Errorf("expected nil but got %v", result)
			}
			result, err = DeserializeInputCommitmentForAll(commitsMap, 3, 3)
			if err == nil {
				t.Errorf("expected error but got none")
			}
			if result != nil {
				t.Errorf("expected nil but got %v", result)
			}
		})
	}
}

func TestSerializeDeserializeCommitments_RoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		commitments [][]byte
	}{
		{
			name:        "Empty commitments",
			commitments: [][]byte{},
		},
		{
			name:        "Single commitment",
			commitments: [][]byte{generateCommitment(0)},
		},
		{
			name: "Multiple commitments",
			commitments: [][]byte{
				generateCommitment(0),
				generateCommitment(1),
				generateCommitment(2),
				generateCommitment(3),
				generateCommitment(4),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized, err := SerializeCommitments(tt.commitments, FrostCommitmentLength)
			if err != nil {
				t.Errorf("SerializeCommitments() unexpected error: %v", err)
				return
			}

			// Deserialize
			result, err := DeserializeCommitments(serialized, len(tt.commitments), FrostCommitmentLength)
			if err != nil {
				t.Errorf("DeserializeCommitments() unexpected error: %v", err)
				return
			}

			// Compare
			if len(result) != len(tt.commitments) {
				t.Errorf("Round trip failed: length = %d, expected %d", len(result), len(tt.commitments))
				return
			}

			for i, commitment := range result {
				if len(commitment) != FrostCommitmentLength {
					t.Errorf("Round trip failed: commitment %d length = %d, expected %d", i, len(commitment), FrostCommitmentLength)
					continue
				}
				for j, b := range commitment {
					if b != tt.commitments[i][j] {
						t.Errorf("Round trip failed: commitment %d byte %d = %d, expected %d", i, j, b, tt.commitments[i][j])
					}
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

type generateDkgR2PackagesFlag uint

const (
	DKG_R2_CORRUPT_PACKAGES_COUNT_INC generateDkgR2PackagesFlag = 1 << iota
	DKG_R2_CORRUPT_PACKAGES_COUNT_DEC
	DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF
	DKG_R2_CORRUPT_VALIDATOR_IDX
	DKG_R2_CORRUPT_PACKAGE_PAYLOAD_INC
	DKG_R2_CORRUPT_PACKAGE_PAYLOAD_DEC
)

func generateDkgR2Packages(validatorsCount uint16, flags generateDkgR2PackagesFlag) map[uint16][]byte {
	r2Packages := make(map[uint16][]byte)
	tmpBuf := make([]byte, 2)

	for fromValidatorIdx := uint16(0); fromValidatorIdx < validatorsCount; fromValidatorIdx++ {
		data := make([]byte, 0)

		// Packages count
		if flags&DKG_R2_CORRUPT_PACKAGES_COUNT_INC != 0 {
			binary.BigEndian.PutUint16(tmpBuf, validatorsCount*2)
		} else if flags&DKG_R2_CORRUPT_PACKAGES_COUNT_DEC != 0 {
			binary.BigEndian.PutUint16(tmpBuf, validatorsCount/2)
		} else if flags&DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF != 0 {
			binary.BigEndian.PutUint16(tmpBuf, validatorsCount)
		} else {
			binary.BigEndian.PutUint16(tmpBuf, validatorsCount-1)
		}
		data = append(data, tmpBuf...)

		// Packages
		for toValidatorIdx := uint16(0); toValidatorIdx < validatorsCount; toValidatorIdx++ {
			if flags&DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF == 0 {
				if toValidatorIdx == fromValidatorIdx {
					continue
				}
			}

			// To validator idx
			if flags&DKG_R2_CORRUPT_VALIDATOR_IDX != 0 {
				binary.BigEndian.PutUint16(tmpBuf, toValidatorIdx>>3)
			} else {
				binary.BigEndian.PutUint16(tmpBuf, toValidatorIdx)
			}
			data = append(data, tmpBuf...)

			// Package payload
			packagePayloadSize := EncryptedFrostDkgR2PackageSize
			if flags&DKG_R2_CORRUPT_PACKAGE_PAYLOAD_INC != 0 {
				packagePayloadSize = EncryptedFrostDkgR2PackageSize + 5
			} else if flags&DKG_R2_CORRUPT_PACKAGE_PAYLOAD_DEC != 0 {
				packagePayloadSize = EncryptedFrostDkgR2PackageSize - 5
			}
			r2Package := make([]byte, packagePayloadSize)
			rand.Read(r2Package)
			data = append(data, r2Package...)
		}

		//
		r2Packages[fromValidatorIdx] = data
	}

	return r2Packages
}

func TestDeserializeDkgR2(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners, 0)
	vsetMask := big.NewInt(0b111)

	resMap, _, _, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if err != nil {
		t.Errorf("Deserialization error: %v", err)
		return
	}

	// Verify resMap
	if len(resMap) != 3 {
		t.Errorf("len(resMap) = %d expected to be %d", len(resMap), 3)
		return
	}
}

func TestDeserializeDkgR2_VSet_1(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners, 0)
	vsetMask := big.NewInt(0b101) // Exclude validator (idx = 1)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `toValidatorIdx is not in VSet`")
		return
	}
}

func TestDeserializeDkgR2_Empty(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	vsetMask := big.NewInt(0b111)
	r2Packages := make(map[uint16][]byte)

	_, _, _, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if err == nil {
		t.Errorf("Expected error `r2Packages is empty`")
		return
	}
}

func TestDeserializeDkgR2_AboveMaxSigners(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners+1, 0)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `incorrect package count`")
		return
	}
}

func TestDeserializeDkgR2_BelowMaxSigners(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, 0)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `incorrect package count`")
		return
	}
}

func TestDeserializeDkgR2_WrongPAckagesCount1(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, DKG_R2_CORRUPT_PACKAGES_COUNT_INC)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `not enough bytes in package`")
		return
	}
}

func TestDeserializeDkgR2_WrongPAckagesCount2(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, DKG_R2_CORRUPT_PACKAGES_COUNT_DEC)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `not enough bytes in package`")
		return
	}
}

func TestDeserializeDkgR2_SendPackageToSelf(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `incorrect package count`")
		return
	}
}

func TestDeserializeDkgR2_CorruptValidatorIdx(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, DKG_R2_CORRUPT_VALIDATOR_IDX)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `incorrect package count`")
		return
	}
}

func TestDeserializeDkgR2_CorruptPackagePayloadSizeInc(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, DKG_R2_CORRUPT_PACKAGE_PAYLOAD_INC)
	vsetMask := big.NewInt(0b111)

	_, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if !isCulprit || culpritId != 0 {
		t.Errorf("Expected isCulprit = true and culpritId = 0, actual values isCulprit = %v, culpritId = %d", isCulprit, culpritId)
		return
	}

	if err == nil {
		t.Errorf("Expected error `incorrect package count`")
		return
	}
}

func TestDeserializeDkgR2_CorruptPackagePayloadSizeDec(t *testing.T) {
	// Build DkgR2 package
	maxSigners := uint16(3)
	r2Packages := generateDkgR2Packages(maxSigners-1, DKG_R2_CORRUPT_PACKAGE_PAYLOAD_DEC)
	vsetMask := big.NewInt(0b111)

	_, _, _, err := DeserializeDkgR2(r2Packages, vsetMask, maxSigners)

	if err == nil {
		t.Errorf("Expected error `not enough bytes in package`")
		return
	}
}
