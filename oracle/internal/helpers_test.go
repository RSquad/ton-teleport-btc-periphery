package helpers

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"strings"
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

type generateDkgR2PackagesFlag uint

const (
	DKG_R2_CORRUPT_PACKAGES_COUNT_INC generateDkgR2PackagesFlag = 1 << iota
	DKG_R2_CORRUPT_PACKAGES_COUNT_DEC
	DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF
	DKG_R2_CORRUPT_TO_VALIDATOR_IDX
	DKG_R2_CORRUPT_PACKAGE_PAYLOAD_INC
	DKG_R2_CORRUPT_PACKAGE_PAYLOAD_DEC
	DKG_R2_CORRUPT_IGNORE_VSET
)

func generateDkgR2Packages(culpritIdx int, vSetMask *big.Int, flags generateDkgR2PackagesFlag) map[uint16][]byte {
	r2Packages := make(map[uint16][]byte)
	tmpBuf := make([]byte, 2)

	maxValidatorsCount := uint16(vSetMask.BitLen())

	packagesCount := uint16(0)
	for i := 0; i < int(maxValidatorsCount); i++ {
		if vSetMask.Bit(i) == 1 {
			packagesCount++
		}
	}

	for fromValidatorIdx := 0; fromValidatorIdx < int(maxValidatorsCount); fromValidatorIdx++ {
		if vSetMask.Bit(fromValidatorIdx) == 0 {
			continue
		}

		data := make([]byte, 0)

		// Packages count
		binary.BigEndian.PutUint16(tmpBuf, packagesCount-1)
		if fromValidatorIdx == culpritIdx {
			if flags&DKG_R2_CORRUPT_PACKAGES_COUNT_INC != 0 {
				binary.BigEndian.PutUint16(tmpBuf, packagesCount*2)
			} else if flags&DKG_R2_CORRUPT_PACKAGES_COUNT_DEC != 0 {
				binary.BigEndian.PutUint16(tmpBuf, packagesCount/2)
			} else if flags&DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF != 0 {
				binary.BigEndian.PutUint16(tmpBuf, packagesCount)
			}
		}
		data = append(data, tmpBuf...)

		// Packages
		for toValidatorIdx := uint16(0); toValidatorIdx < maxValidatorsCount; toValidatorIdx++ {
			if fromValidatorIdx == culpritIdx {
				if flags&DKG_R2_CORRUPT_IGNORE_VSET == 0 {
					if vSetMask.Bit(int(toValidatorIdx)) == 0 {
						continue
					}
				}
			} else {
				if vSetMask.Bit(int(toValidatorIdx)) == 0 {
					continue
				}
			}

			if fromValidatorIdx == culpritIdx {
				if flags&DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF == 0 {
					if int(toValidatorIdx) == fromValidatorIdx {
						continue
					}
				}
			} else {
				if int(toValidatorIdx) == fromValidatorIdx {
					continue
				}
			}

			// To validator idx
			binary.BigEndian.PutUint16(tmpBuf, toValidatorIdx)
			if fromValidatorIdx == culpritIdx {
				if flags&DKG_R2_CORRUPT_TO_VALIDATOR_IDX != 0 {
					binary.BigEndian.PutUint16(tmpBuf, (toValidatorIdx+1)<<4)
				}
			}
			data = append(data, tmpBuf...)

			// Package payload
			packagePayloadSize := EncryptedFrostDkgR2PackageSize
			if fromValidatorIdx == culpritIdx {
				if flags&DKG_R2_CORRUPT_PACKAGE_PAYLOAD_INC != 0 {
					packagePayloadSize = EncryptedFrostDkgR2PackageSize + 5
				} else if flags&DKG_R2_CORRUPT_PACKAGE_PAYLOAD_DEC != 0 {
					packagePayloadSize = EncryptedFrostDkgR2PackageSize - 5
				}
			}

			r2Package := make([]byte, packagePayloadSize)
			rand.Read(r2Package)
			data = append(data, r2Package...)
		}

		r2Packages[uint16(fromValidatorIdx)] = data
	}

	return r2Packages
}

func TestDeserializeDkgR2(t *testing.T) {
	tests := []struct {
		vsetMask      *big.Int
		culpritIdx    int
		flags         generateDkgR2PackagesFlag
		expectedError string
	}{
		{
			vsetMask:      big.NewInt(0b11111111),
			culpritIdx:    -1,
			flags:         0,
			expectedError: "",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    -1,
			flags:         0,
			expectedError: "",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_IGNORE_VSET,
			expectedError: "incorrect package size",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_TO_VALIDATOR_IDX,
			expectedError: "toValidatorIdx is not in VSet",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_PACKAGES_COUNT_INC,
			expectedError: "incorrect package size",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_PACKAGES_COUNT_DEC,
			expectedError: "incorrect package size",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_SEND_PACKAGE_TO_SELF,
			expectedError: "toValidatorIdx 2 is the same as fromValidatorIdx 2",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_PACKAGE_PAYLOAD_INC,
			expectedError: "incorrect package size",
		},
		{
			vsetMask:      big.NewInt(0b11111101),
			culpritIdx:    2,
			flags:         DKG_R2_CORRUPT_PACKAGE_PAYLOAD_DEC,
			expectedError: "incorrect package size",
		},
	}

	for ii, tt := range tests {
		r2Packages := generateDkgR2Packages(tt.culpritIdx, tt.vsetMask, tt.flags)

		resMap, isCulprit, culpritId, err := DeserializeDkgR2(r2Packages, tt.vsetMask)

		if tt.culpritIdx >= 0 {
			if !isCulprit || culpritId != uint16(tt.culpritIdx) {
				t.Errorf("Expected isCulprit = true and culpritId = %d, actual values isCulprit = %v, culpritId = %d", tt.culpritIdx, isCulprit, culpritId)
				continue
			}
		}

		if len(tt.expectedError) == 0 {
			if err != nil {
				t.Errorf("Deserialization error: %v", err)
				continue
			}
		} else {
			if err == nil {
				t.Errorf("Expected error `%s`", tt.expectedError)
				continue
			}

			if strings.Contains(err.Error(), tt.expectedError) == false {
				t.Errorf("Expected error `%s`, but got `%s`", tt.expectedError, err.Error())
			}

			continue
		}

		// Verify resMap
		if len(resMap) != len(r2Packages) {
			t.Errorf("len(resMap) = %d expected to be %d, ii = %d", len(resMap), len(r2Packages), ii)
			return
		}
	}
}
