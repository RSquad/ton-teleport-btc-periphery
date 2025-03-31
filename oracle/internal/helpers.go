package helpers

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/rsquad/ton-teleport-btc-periphery/frost"
)

// Helpers

func ConvertMapToFrostPackages(origMap map[string][]byte) (frostMap map[frost.Identifier]frost.Package) {
	frostMap = make(map[frost.Identifier]frost.Package)
	for k, v := range origMap {
		fmt.Println("ConvertMapToFrostPackages: k = '", k, "'")
		id, _ := frost.DecodeIdentifier(k)
		frostMap[*id] = frost.NewPackage(v)
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
	case 114:
		return "package already sent"
	case 127:
		return "R1 is not completed yet"
	case 128:
		return "R2 is not completed yet"
	case 145:
		return "Commitments threshold is reached"
	case 147:
		return "R1 is already completed"
	case 150:
		return "Coordinator balance is not enough to continue"
	case 161:
		return "Unauthorized validator"
	default:
		return fmt.Sprintf("Unknown error: %d", exitCode)
	}
}

func ValidatorIdxToFrost(validatorIdx uint16) []byte {
	validatorIdx |= 0x80
	return frost.GetIdentifier(validatorIdx)
}
