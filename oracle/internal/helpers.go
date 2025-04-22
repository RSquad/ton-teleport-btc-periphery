package helpers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

const TvmExitCodeDifferentPubkeyPackages = 152
const FrostDkgR2PackageSize = 37 /*FROST R2 package to single validator*/
const EncryptedFrostDkgR2PackageSize = 24 /*nonce for encryption*/ + 16 /*encryption header*/ + FrostDkgR2PackageSize
const SizeOfSingleDkgR2Package = 2 /*ToValidatorId*/ + EncryptedFrostDkgR2PackageSize

// Helpers

func DeserializeDkgR1(origMap map[uint16][]byte) (map[frost.Identifier]frost.Package, map[uint16][]byte, uint16, error) {
	frostMap := make(map[frost.Identifier]frost.Package)
	pubKeysMap := make(map[uint16][]byte)

	for validatorIdx, pkgData := range origMap {
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

	for toValidatorIdx, r2pkg := range r2Packages {
		// Serialize validator idx
		binary.BigEndian.PutUint16(tmpBuf, toValidatorIdx)
		serializedData = append(serializedData, tmpBuf...)

		// Serialize r2pkg
		serializedData = append(serializedData, r2pkg...)
	}

	return serializedData
}

func DeserializeDkgR2(r2Packages map[uint16][]byte /*map[FROM]data*/, vsetMask *big.Int, maxSigners uint16) (
	map[uint16]map[uint16][]byte, /*map[FROM]map[TO]data*/
	bool, /*is culprit was found*/
	uint16, /*culprit*/
	error) {
	deserializedData := make(map[uint16]map[uint16][]byte)

	for fromValidatorIdx, serializedToPkgs := range r2Packages {
		toValidatorData := make(map[uint16][]byte)

		readOffset := 0
		bytesLeft := len(serializedToPkgs)

		// Packages count
		if bytesLeft < 2 {
			return nil, true, fromValidatorIdx, errors.New("not enough bytes in package")
		}

		packagesCount := binary.BigEndian.Uint16(serializedToPkgs[readOffset : readOffset+2])
		readOffset += 2

		for range packagesCount {
			if bytesLeft < SizeOfSingleDkgR2Package {
				return nil, true, fromValidatorIdx, errors.New("not enough bytes in package")
			}

			// To validator idx
			toValidatorIdx := binary.BigEndian.Uint16(serializedToPkgs[readOffset : readOffset+2])
			readOffset += 2

			// Check if toValidatorIdx is unique
			_, exists := toValidatorData[toValidatorIdx]
			if exists {
				return nil, true, fromValidatorIdx, fmt.Errorf("validator ID %d is not unique", toValidatorIdx)
			}

			toValidatorData[toValidatorIdx] = serializedToPkgs[readOffset : readOffset+EncryptedFrostDkgR2PackageSize]
			readOffset += EncryptedFrostDkgR2PackageSize
		}

		// Check toValidatorData. All and only the validator indexes from VSet must be in toValidatorData (exept fromValidatorIdx)
		count := uint(0)
		for toValidatorIdx := range toValidatorData {
			count += vsetMask.Bit(int(toValidatorIdx))
		}

		if count != uint(maxSigners-1 /*fromValidatorIdx*/) {
			return nil, true, fromValidatorIdx, errors.New("validator did not send R2 packages for some validators from the VSet")
		}

		deserializedData[fromValidatorIdx] = toValidatorData
	}

	return deserializedData, false, 0, nil
}

func ConvertMapToFrostPackages(origMap map[uint16][]byte) (frostMap map[frost.Identifier]frost.Package) {
	frostMap = make(map[frost.Identifier]frost.Package)
	for k, v := range origMap {
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
	case 147:
		return "R1 is already completed"
	case 150:
		return "Coordinator balance is not enough to continue"
	case 151:
		return "Culprit not found"
	case 161:
		return "Unauthorized validator"
	case 166:
		return "Pegout is not expired"
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
