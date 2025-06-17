package helpers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"sort"
	"strconv"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"golang.org/x/exp/constraints"
)

const (
	ErrDkgClosed                       = 112
	ErrPackageAlreadyExist             = 114
	ErrCommitmentsThresholdReached     = 145
	ErrDkgAlreadyExecuted              = 146
	TvmExitCodeDifferentPubkeyPackages = 152
	ErrClaimAlreadyExists              = 160
	ErrPegoutIsNotExpired              = 166
	DifferentPegoutSignatures          = 168
	ErrDkgExpired                      = 182
	FrostDkgR2PackageSize              = 37 /*FROST R2 package to single validator*/
	FrostDkgR2PackageForEncryptionSize = 8 /*DKG until*/ + 2 /*from validator idx*/ + FrostDkgR2PackageSize
	EncryptedFrostDkgR2PackageSize     = 24 /*nonce for encryption*/ + 16 /*encryption header*/ + FrostDkgR2PackageForEncryptionSize
	SizeOfSingleDkgR2Package           = 2 /*ToValidatorId*/ + EncryptedFrostDkgR2PackageSize
	FrostCommitmentLength              = 71
)

// Helpers
func ExtractSortedKeysFromMap[K constraints.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func DeserializeDkgR1(origMap map[uint16][]byte) (map[frost.Identifier]frost.Package, map[uint16][]byte, uint16, error) {
	frostMap := make(map[frost.Identifier]frost.Package)
	pubKeysMap := make(map[uint16][]byte)

	for _, validatorIdx := range ExtractSortedKeysFromMap(origMap) {
		pkgData := origMap[validatorIdx]

		if len(pkgData) < 33 { // 32 - public key; at least 1 byte of data
			return nil, nil, validatorIdx, errors.New("wrong package len")
		}

		pubKeyX25519 := pkgData[:32]
		frostPackage := pkgData[32:]

		pubKeysMap[validatorIdx] = pubKeyX25519
		frostMap[ValidatorIdxToFrost(validatorIdx)] = frost.NewPackage(frostPackage)
	}

	return frostMap, pubKeysMap, 0, nil
}

func SerializeR2Packages(r2Packages map[uint16][]byte) []byte {
	serializedData := []byte{}
	tmpBuf := make([]byte, 2)

	// Serialize packages count
	binary.BigEndian.PutUint16(tmpBuf, uint16(len(r2Packages)))
	serializedData = append(serializedData, tmpBuf...)

	for _, toValidatorIdx := range ExtractSortedKeysFromMap(r2Packages) {
		r2pkg := r2Packages[toValidatorIdx]

		// Serialize validator idx
		binary.BigEndian.PutUint16(tmpBuf, toValidatorIdx)
		serializedData = append(serializedData, tmpBuf...)

		// Serialize r2pkg
		serializedData = append(serializedData, r2pkg...)
	}

	return serializedData
}

func DeserializeDkgR2(r2Packages map[uint16][]byte /*map[FROM]data*/, vSetMask *big.Int) (
	map[uint16]map[uint16][]byte, /*map[FROM]map[TO]data*/
	bool, /*is culprit was found*/
	uint16, /*culprit*/
	error,
) {
	deserializedData := make(map[uint16]map[uint16][]byte)

	if len(r2Packages) == 0 {
		return nil, false, 0, errors.New("r2Packages is empty")
	}

	// We need to iterate over r2Packages in deterministic order (which is not guaranteed for map[])
	for _, fromValidatorIdx := range ExtractSortedKeysFromMap(r2Packages) {
		toValidatorData := make(map[uint16][]byte)
		serializedToPkgs := r2Packages[fromValidatorIdx]
		packagesSize := len(serializedToPkgs)

		if vSetMask.Bit(int(fromValidatorIdx)) == 0 {
			return nil, true, fromValidatorIdx, fmt.Errorf("fromValidatorIdx %d in not in VSet", fromValidatorIdx)
		}

		// Packages count
		if packagesSize < 2 {
			return nil, true, fromValidatorIdx, errors.New("not enough bytes in package")
		}

		packagesCount := int(binary.BigEndian.Uint16(serializedToPkgs[0:2]))

		expectedSize := packagesCount * (2 /*validator idx*/ + EncryptedFrostDkgR2PackageSize)
		if (packagesSize - 2) != expectedSize {
			return nil, true, fromValidatorIdx, fmt.Errorf("incorrect package size, expected size = %d, actual size = %d", expectedSize, packagesSize-2)
		}

		readOffset := 2

		for range packagesCount {
			// To validator idx
			toValidatorIdx := binary.BigEndian.Uint16(serializedToPkgs[readOffset : readOffset+2])
			readOffset += 2

			// Check VSet
			if vSetMask.Bit(int(toValidatorIdx)) == 0 {
				return nil, true, fromValidatorIdx, fmt.Errorf("toValidatorIdx is not in VSet. fromValidatorIdx %d, toValidatorIdx %d", fromValidatorIdx, toValidatorIdx)
			}

			// Check if toValidatorIdx != fromValidatorIdx
			if toValidatorIdx == fromValidatorIdx {
				return nil, true, fromValidatorIdx, fmt.Errorf("toValidatorIdx %d is the same as fromValidatorIdx %d", toValidatorIdx, fromValidatorIdx)
			}

			// Check if toValidatorIdx is unique
			_, exists := toValidatorData[toValidatorIdx]
			if exists {
				return nil, true, fromValidatorIdx, fmt.Errorf("validator ID %d is not unique", toValidatorIdx)
			}

			toValidatorData[toValidatorIdx] = serializedToPkgs[readOffset : readOffset+EncryptedFrostDkgR2PackageSize]
			readOffset += EncryptedFrostDkgR2PackageSize
		}

		// Count vSet size
		vSetSize := 0
		for i := 0; i < vSetMask.BitLen(); i++ {
			if vSetMask.Bit(i) == 1 {
				vSetSize++
			}
		}

		if len(toValidatorData) != int(vSetSize-1 /*ourself*/) {
			return nil, true, fromValidatorIdx, fmt.Errorf("valid packages count %d from Validator %d is not equal to vSetSize: expected %d", len(toValidatorData), fromValidatorIdx, vSetSize)
		}

		deserializedData[fromValidatorIdx] = toValidatorData
	}

	return deserializedData, false, 0, nil
}

func SerializeCommitments(commitments [][]byte, expectedLength int) ([]byte, error) {
	serialized := []byte{}
	serialized = append(serialized, byte(len(commitments)))
	for _, commitment := range commitments {
		if len(commitment) != expectedLength {
			return nil, fmt.Errorf("commitment length %d is not equal to %d", len(commitment), FrostCommitmentLength)
		}
		serialized = append(serialized, commitment...)
	}
	return serialized, nil
}

func DeserializeCommitments(serialized []byte, expectedCount int, expectedLength int) ([][]byte, error) {
	if len(serialized) == 0 {
		return nil, errors.New("serialized commitments data is empty")
	}

	commitments := [][]byte{}
	commitmentsCount := int(serialized[0])

	if commitmentsCount != expectedCount {
		return nil, fmt.Errorf("incorrect number of commitments: expected %d, got %d", expectedCount, commitmentsCount)
	}

	if len(serialized) < 1+commitmentsCount*expectedLength {
		return nil, fmt.Errorf("insufficient data: cannot read commitments (expected %d bytes, have %d)", 1+commitmentsCount*expectedLength, len(serialized))
	}

	for i := 0; i < commitmentsCount; i++ {
		commitment := make([]byte, expectedLength)
		copy(commitment, serialized[1+i*expectedLength:1+(i+1)*expectedLength])
		commitments = append(commitments, commitment)
	}

	return commitments, nil
}

// Deserializes only 1 commitment with index `inputIndex` for every validator
func DeserializeInputCommitmentForAll(
	validatorCommitments map[uint16][]byte, // all commitments for all validators
	totalInputs int, // total number of inputs in pegout transaction - used for validation
	inputIndex int, // pegout transaction input index for which we need to deserialize commitments
) (map[uint16][]byte, error) {
	if inputIndex < 0 || inputIndex >= totalInputs {
		return nil, fmt.Errorf("inputIndex is out of range: %d", inputIndex)
	}

	commitmentsMap := make(map[uint16][]byte)
	// for each validator, deserialize all commitments and return the commitment for the inputIndex
	for validatorIdx, serializedCommitments := range validatorCommitments {
		commitments, err := DeserializeCommitments(serializedCommitments, totalInputs, FrostCommitmentLength)
		if err != nil {
			return nil, err
		}
		commitmentsMap[validatorIdx] = commitments[inputIndex]
	}
	return commitmentsMap, nil
}

func ConvertMapToFrostPackages(origMap map[uint16][]byte) (frostMap map[frost.Identifier]frost.Package) {
	frostMap = make(map[frost.Identifier]frost.Package)
	for _, k := range ExtractSortedKeysFromMap(origMap) {
		v := origMap[k]
		id := ValidatorIdxToFrost(k)
		frostMap[id] = frost.NewPackage(v)
	}
	return
}

func CalcMinSigners(maxSigners uint16) (uint16, error) {
	if maxSigners < 2 {
		return 0, errors.New("maxSigners must be greater than 1")
	}
	minSigners := uint16(math.Floor(float64(maxSigners) * 2 / 3))
	return max(minSigners, 2), nil
}

// ExtractExitCode extracts the exitcode value from a TON VM error log
func ExtractExitCode(errorLog string) (int, error) {
	exitCodePattern := regexp.MustCompile(`exitcode=(\d+)`)
	matches := exitCodePattern.FindStringSubmatch(errorLog)

	if len(matches) >= 2 {
		exitCode, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, errors.New("failed to parse exitcode")
		}
		return exitCode, nil
	}

	// If we reach here, we couldn't find an exitcode
	return 0, errors.New("exitcode not found in error log")
}

func HandleTvmError(tvmError error) string {
	exitCode, err := ExtractExitCode(tvmError.Error())
	if err != nil {
		return tvmError.Error()
	}

	switch exitCode {
	case 112:
		return "Dkg closed"
	case 113:
		return "invalid signature"
	case 114:
		return "package already sent"
	case 117:
		return "signature exists"
	case 127:
		return "R1 is not completed yet"
	case 128:
		return "R2 is not completed yet"
	case 135:
		return "pegout not found"
	case 145:
		return "Commitments threshold is reached"
	case 146:
		return "Dkg already executed"
	case 147:
		return "R1 is already completed"
	case 150:
		return "Coordinator balance is not enough to continue"
	case 151:
		return "Culprit not found"
	case 152:
		return "Different pubkey packages"
	case 160:
		return "Claim already exists"
	case 161:
		return "Unauthorized validator"
	case 162:
		return "Not enough validators"
	case 166:
		return "Pegout is not expired"
	case 168:
		return "Different pegout signatures"
	case 171:
		return "Pegout id does not match expected pegout to sign"
	case 180:
		return "Invalid pegout timestamp"
	case 181:
		return "Invalid session signature"
	case 182:
		return "DKG expired"
	default:
		return fmt.Sprintf("Unknown error: %d", exitCode)
	}
}

func ValidatorIdxToFrost(validatorIdx uint16) frost.Identifier {
	validatorIdx |= 0x80
	return frost.GetIdentifier(validatorIdx)
}

func FrostToValidatorIdx(frostIdentificator frost.Identifier) uint16 {
	return binary.BigEndian.Uint16(frostIdentificator[30:32]) & (^uint16(0x80))
}

func ParseIntWithDefaultVal(str string, defaultValue int64, name string) int64 {
	value := defaultValue

	if len(str) > 0 {
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			logger.Log.Warn().Msgf("Failed to parse %s value `%s`. Default value of %ds will be used.", name, str, defaultValue)
		} else {
			value = val
		}
	}

	return value
}
