package ton

import (
    "context"
    "crypto/ed25519"

    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/ton/wallet"
)

type WalletContract struct {
    *wallet.Wallet
    api *ton.APIClient
    ctx context.Context
}

func NewWalletContract(api *ton.APIClient, secret ed25519.PrivateKey, ctx context.Context) (*WalletContract, error) {
    w, err := wallet.FromPrivateKey(api, secret, wallet.V4R2)
    if err != nil {
        return nil, err
    }
    return &WalletContract{
        Wallet: w,
        api:    api,
        ctx:    ctx,
    }, nil
}
