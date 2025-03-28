package validator

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/keystore"
)

type KeyInfo struct {
	KeyID     []byte
	VsetIdx   uint16
	PublicKey []byte
}

func NewKeyInfo(keyID []byte, publicKey []byte) KeyInfo {
	return KeyInfo{
		KeyID:     keyID,
		VsetIdx:   0,
		PublicKey: publicKey,
	}
}

type SignerType int

const (
	_ SignerType = iota
	SIGNER_VALIDATOR
	SIGNER_ORACLE
)

type Validator struct {
	standaloneMode       bool
	standalonePrivateKey []byte
	standalonePublicKey  []byte
	useOracleSign        bool
	validatorConsole     *ValidatorConsole
}

func NewValidator(cfg *cfg.Cfg) (*Validator, error) {
	var standalonePublicKey []byte = nil
	var standalonePrivateKey []byte = nil
	var err error = nil
	var console *ValidatorConsole
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
		console = NewValidatorConsole(
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
		useOracleSign:        false,
		validatorConsole:     console,
	}, nil
}

func (v *Validator) FindKeyInfo(vset coordinator.VSet, keystore *keystore.Keystore) (*KeyInfo, error) {
	if v.standaloneMode {
		// find vset idx by public key in dkg.VSet
		for idx, pubkey := range vset {
			if bytes.Equal(pubkey, v.standalonePublicKey) {
				return &KeyInfo{
					KeyID:     v.standalonePublicKey,
					VsetIdx:   idx,
					PublicKey: v.standalonePublicKey,
				}, nil
			}
		}
		return nil, errors.New("validator key not found")
	}

	// Try to get keys from validator console
	validatorKeys, err := v.validatorConsole.GetValidatorKeys()
	if err != nil {
		return nil, err
	}

	for _, keyInfo := range validatorKeys {
		for idx, pubkey := range vset {
			if bytes.Equal(pubkey, keyInfo.PublicKey) {
				keyInfo.VsetIdx = idx
				return &keyInfo, nil
			}
		}
	}

	// If the key was not found in the validator key set, then try to find it in the oracle session key set
	for idx, pubkey := range vset {
		secret := (*keystore).LoadSecret(pubkey)
		if secret != nil {
			return &KeyInfo{
				KeyID:     pubkey,
				VsetIdx:   idx,
				PublicKey: pubkey,
			}, nil
		}
	}

	return nil, errors.New("No key was found")
}

func (v *Validator) GetSigner(keyID []byte, keystore *keystore.Keystore, signerType SignerType) signer.Signer {
	if v.standaloneMode {
		return signer.NewKeySigner(hex.EncodeToString(v.standalonePrivateKey))
	}

	if signerType == SIGNER_ORACLE {
		return NewOracleSigner(keyID, keystore)
	}

	// SIGNER_VALIDATOR
	return NewValidatorSigner(v.validatorConsole, hex.EncodeToString(keyID))
}
