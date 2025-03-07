package validator

import (
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/oracle/internal/cfg"
)

type ValidatorKey struct {
	ValidatorID  int
	ValidatorIdx int
	ValidatorKey []byte
}

type Validator struct{}

func NewValidator(cfg *cfg.Cfg) *Validator {
	return &Validator{}
}

func (v *Validator) GetValidatorKey(dkg *coordinator.DKG) (*ValidatorKey, error) {
	panic("not implemented")
}

func (v *Validator) GetSigner(validatorID int) *signer.Signer {
	panic("not implemented")
}
