package validator

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	validatorconsole "github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/validator_console"
)

type ValidatorKeyInfo struct {
	KeyID     []byte
	VsetIdx   uint16
	PublicKey []byte
}

type Validator struct {
	standaloneMode       bool
	standalonePrivateKey []byte
	standalonePublicKey  []byte
	validatorConsole     validatorconsole.ValidatorConsole
}

func NewValidator(cfg *cfg.Cfg) (*Validator, error) {
	var standalonePublicKey []byte = nil
	var standalonePrivateKey []byte = nil
	var err error = nil
	var console validatorconsole.ValidatorConsole
	if cfg.StandaloneMode {
		standalonePublicKey, err = hex.DecodeString(cfg.Pubkey)
		if err != nil {
			return nil, err
		}
		standalonePrivateKey, err = hex.DecodeString(cfg.Secret)
		if err != nil {
			return nil, err
		}
	} else {
		console = *validatorconsole.NewValidatorConsole(
			cfg.ValidatorEngineConsolePath,
			cfg.ServerPublicKeyPath,
			cfg.ClientPrivateKeyPath,
			cfg.ValidatorServerAddr,
		)
	}
	return &Validator{
		standaloneMode:       cfg.StandaloneMode,
		standalonePublicKey:  standalonePublicKey,
		standalonePrivateKey: standalonePrivateKey,
		validatorConsole:     console,
	}, nil
}

func (v *Validator) FindKeyInfo(vset coordinator.VSet) (*ValidatorKeyInfo, error) {
	if v.standaloneMode {
		// find vset idx by public key in dkg.VSet
		for idx, pubkey := range vset {
			if bytes.Equal(pubkey, v.standalonePublicKey) {
				return &ValidatorKeyInfo{
					KeyID:     v.standalonePublicKey,
					VsetIdx:   idx,
					PublicKey: v.standalonePublicKey,
				}, nil
			}
		}
		return nil, errors.New("validator key not found")
	} else {
		vals, err := v.validatorConsole.GetValidatorKeys()
		if err != nil {
			return nil, err
		}
		for _, val := range vals {
			for idx, pubkey := range vset {
				if bytes.Equal(pubkey, val.ValidatorKey) {
					return &ValidatorKeyInfo{
						KeyID:     []byte(val.ValidatorId),
						VsetIdx:   idx,
						PublicKey: val.ValidatorKey,
					}, nil
				}
			}
		}
	}
	return nil, errors.New("validator key not found")
}

func (v *Validator) GetSigner(keyID []byte) *signer.Signer {
	if v.standaloneMode {
		return signer.New(hex.EncodeToString(v.standalonePrivateKey))
	}
	// else {
	// 	return v.validatorConsole.NewValidatorSigner(hex.EncodeToString(keyID))
	// }
	panic("not implemented")
}
