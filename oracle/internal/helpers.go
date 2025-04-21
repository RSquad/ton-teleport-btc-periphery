package helpers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

const TvmExitCodeDifferentPubkeyPackages = 152

// Helpers

func FromFrostAndPubKeyPkg(origMap map[uint16][]byte) (map[frost.Identifier]frost.Package, map[uint16][]byte, uint16, error) {
	frostMap := make(map[frost.Identifier]frost.Package)
	pubKeysMap := make(map[uint16][]byte)

	for validatorIdx, pkgData := range origMap {
		if len(pkgData) < 33 { /*33 = 32 [public key] + 1[The length of the frost package is expected to be at least 1 byte]*/
			return nil, nil, validatorIdx, errors.New("wrong package len") // TODO: add custom culprit error
		}

		pubKeyX25519 := pkgData[:32]
		frostPackage := pkgData[32:]

		pubKeysMap[validatorIdx] = pubKeyX25519
		frostMap[ValidatorIdxToFrost(validatorIdx)] = frost.NewPackage(frostPackage)
	}

	return frostMap, pubKeysMap, 0, nil
}

func FromFrostPkg(origMap map[uint16][]byte) (frostMap map[frost.Identifier]frost.Package) {
	frostMap = make(map[frost.Identifier]frost.Package)
	for validatorIdx, v := range origMap {
		id := ValidatorIdxToFrost(validatorIdx)
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
