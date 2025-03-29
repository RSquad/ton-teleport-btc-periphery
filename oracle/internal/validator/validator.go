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

type Validator struct {
	standaloneMode       bool
	standalonePrivateKey []byte
	standalonePublicKey  []byte
	validatorConsole     *ValidatorConsole
	sessionSigner        *SessionSigner
}

func NewValidator(cfg *cfg.Cfg) (*Validator, error) {
	var standalonePublicKey []byte = nil
	var standalonePrivateKey []byte = nil
	var err error = nil
	var console *ValidatorConsole
	var sessionSigner *SessionSigner

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

	sessionSigner = NewSessionSigner(cfg.KeystorePath)

	return &Validator{
		standaloneMode:       cfg.StandaloneMode,
		standalonePublicKey:  standalonePublicKey,
		standalonePrivateKey: standalonePrivateKey,
		validatorConsole:     console,
		sessionSigner:        sessionSigner,
	}, nil
}

func (v *Validator) FindKeyInfo(vset coordinator.VSet) (*KeyInfo, error) {
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

	return nil, errors.New("no key was found")
}

func (v *Validator) GetSigner(keyID []byte, keystore *keystore.Keystore) signer.Signer {
	if v.standaloneMode {
		return signer.NewKeySigner(hex.EncodeToString(v.standalonePrivateKey))
	}

	return NewValidatorSigner(v.validatorConsole, hex.EncodeToString(keyID))
}

func (v *Validator) GetSessionSigner() *SessionSigner {
	return v.sessionSigner
}
