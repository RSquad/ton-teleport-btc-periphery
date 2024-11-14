package jwv4r2contract

import (
	"context"
	"crypto/ed25519"

	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

type JWV4R2Contract struct {
	*wallet.Wallet
	api *ton.APIClient
	ctx context.Context
}

func NewJWV4R2Contract(api *ton.APIClient, secret ed25519.PrivateKey, ctx context.Context) (*JWV4R2Contract, error) {
	w, err := wallet.FromPrivateKey(api, secret, wallet.V4R2)
	if err != nil {
		return nil, err
	}
	return &JWV4R2Contract{
		Wallet: w,
		api:    api,
		ctx:    ctx,
	}, nil
}
