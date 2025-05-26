package frost

import (
	"fmt"
)

type LibFrostError int32

const (
	// Error codes matching Rust FROST error enum
	ErrInvalidMinSigners                LibFrostError = -1
	ErrInvalidMaxSigners                LibFrostError = -2
	ErrInvalidCoefficients              LibFrostError = -3
	ErrMalformedIdentifier              LibFrostError = -4
	ErrDuplicatedIdentifier             LibFrostError = -5
	ErrUnknownIdentifier                LibFrostError = -6
	ErrIncorrectNumberOfIdentifiers     LibFrostError = -7
	ErrMalformedSigningKey              LibFrostError = -8
	ErrMalformedVerifyingKey            LibFrostError = -9
	ErrMalformedSignature               LibFrostError = -10
	ErrInvalidSignature                 LibFrostError = -11
	ErrDuplicatedShares                 LibFrostError = -12
	ErrIncorrectNumberOfShares          LibFrostError = -13
	ErrIdentityCommitment               LibFrostError = -14
	ErrMissingCommitment                LibFrostError = -15
	ErrIncorrectCommitment              LibFrostError = -16
	ErrIncorrectNumberOfCommitments     LibFrostError = -17
	ErrInvalidSignatureShare            LibFrostError = -18
	ErrInvalidSecretShare               LibFrostError = -19
	ErrPackageNotFound                  LibFrostError = -20
	ErrIncorrectNumberOfPackages        LibFrostError = -21
	ErrIncorrectPackage                 LibFrostError = -22
	ErrDKGNotSupported                  LibFrostError = -23
	ErrInvalidProofOfKnowledge          LibFrostError = -24
	ErrFieldError                       LibFrostError = -25
	ErrGroupError                       LibFrostError = -26
	ErrInvalidCoefficient               LibFrostError = -27
	ErrIdentifierDerivationNotSupported LibFrostError = -28
	ErrSerializationError               LibFrostError = -29
	ErrDeserializationError             LibFrostError = -30

	// Other
	Success      LibFrostError = 0
	NullArgument LibFrostError = -126
	Unknown      LibFrostError = -127
)

// Error messages corresponding to FROST errors
var frostErrorMessages = map[LibFrostError]string{
	ErrInvalidMinSigners:                "min_signers must be at least 2 and not larger than max_signers",
	ErrInvalidMaxSigners:                "max_signers must be at least 2",
	ErrInvalidCoefficients:              "coefficients must have min_signers-1 elements",
	ErrMalformedIdentifier:              "malformed identifier is unserializable",
	ErrDuplicatedIdentifier:             "duplicated identifier",
	ErrUnknownIdentifier:                "unknown identifier",
	ErrIncorrectNumberOfIdentifiers:     "incorrect number of identifiers",
	ErrMalformedSigningKey:              "malformed signing key encoding",
	ErrMalformedVerifyingKey:            "malformed verifying key encoding",
	ErrMalformedSignature:               "malformed signature encoding",
	ErrInvalidSignature:                 "invalid signature",
	ErrDuplicatedShares:                 "duplicated shares provided",
	ErrIncorrectNumberOfShares:          "incorrect number of shares",
	ErrIdentityCommitment:               "commitment equals the identity",
	ErrMissingCommitment:                "the signing package must contain the participant's commitment",
	ErrIncorrectCommitment:              "the participant's commitment is incorrect",
	ErrIncorrectNumberOfCommitments:     "incorrect number of commitments",
	ErrInvalidSignatureShare:            "invalid signature share",
	ErrInvalidSecretShare:               "invalid secret share",
	ErrPackageNotFound:                  "round 1 package not found for round 2 participant",
	ErrIncorrectNumberOfPackages:        "incorrect number of packages",
	ErrIncorrectPackage:                 "the incorrect package was specified",
	ErrDKGNotSupported:                  "the ciphersuite does not support DKG",
	ErrInvalidProofOfKnowledge:          "the proof of knowledge is not valid",
	ErrFieldError:                       "error in scalar field",
	ErrGroupError:                       "error in elliptic curve group",
	ErrInvalidCoefficient:               "invalid coefficient",
	ErrIdentifierDerivationNotSupported: "the ciphersuite does not support deriving identifiers from strings",
	ErrSerializationError:               "error serializing value",
	ErrDeserializationError:             "error deserializing value",
	Success:                             "success",
	NullArgument:                        "null argument",
	Unknown:                             "unknonwn",
}

type CulpritInfo struct {
	Id      *Identifier
	ErrCode LibFrostError
}

func NewCulpritInfo(culpritData []byte, errCode LibFrostError) *CulpritInfo {
	var id *Identifier = nil
	if errCode == ErrInvalidSignatureShare || errCode == ErrInvalidSecretShare || errCode == ErrInvalidProofOfKnowledge {
		copy(id[:], culpritData)
	}

	return &CulpritInfo{
		Id:      id,
		ErrCode: errCode,
	}
}

// Error implements the error interface for FrostError
func Error(code int32) error {
	if msg, ok := frostErrorMessages[LibFrostError(code)]; ok {
		return fmt.Errorf("frost error: %s, code %d", msg, code)
	}
	return fmt.Errorf("unknown FROST error code: %d", code)
}
