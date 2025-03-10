package validator

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
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
}

func NewValidator(cfg *cfg.Cfg) (*Validator, error) {
	var standalonePublicKey []byte = nil
	var standalonePrivateKey []byte = nil
	var err error = nil
	if cfg.StandaloneMode {
		standalonePublicKey, err = hex.DecodeString(cfg.Pubkey)
		if err != nil {
			return nil, err
		}
		standalonePrivateKey, err = hex.DecodeString(cfg.Secret)
		if err != nil {
			return nil, err
		}
	}
	return &Validator{
		standaloneMode:       cfg.StandaloneMode,
		standalonePublicKey:  standalonePublicKey,
		standalonePrivateKey: standalonePrivateKey,
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
	}
	panic("not implemented")
}

func (v *Validator) GetSigner(keyID []byte) *signer.Signer {
	if v.standaloneMode {
		return signer.New(hex.EncodeToString(v.standalonePrivateKey))
	}
	panic("not implemented")
}
