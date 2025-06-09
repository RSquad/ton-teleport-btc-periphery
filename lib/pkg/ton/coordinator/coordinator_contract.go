//go:generate moq -out coordinator_contract_mock.go . Coordinator

package coordinator

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/signer"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	tonutils "github.com/xssnick/tonutils-go/ton"
)

type Coordinator interface {
	GetDkg(block *tonutils.BlockIDExt) (*DKG, error)
	GetPrevDKG() (*DKG, error)
	GetUnsignedPegouts() ([]PegoutRecord, error)
	GetStorage(block *tonutils.BlockIDExt) (Storage, error)

	SendStartDKG() (*tlb.Transaction, error)
	SendRound1(
		validatorIdx uint16,
		dkgUntil int64,
		round1Package []byte,
		r2PublicX25519 *[32]byte,
	) (*tlb.Transaction, error)

	SendRound2(
		validatorIdx uint16,
		dkgUntil int64,
		round2Packages []byte,
	) (*tlb.Transaction, error)

	SendDKGClaim(
		validatorIdx uint16,
		dkgUntil int64,
		culpritIdx uint16,
	) (*tlb.Transaction, error)

	SendPubkeyPackage(
		validatorIdx uint16,
		dkgUntil int64,
		sessionSigner signer.Signer,
		pubkeyPackage []byte,
	) (*tlb.Transaction, error)

	SendCommitments(
		pegoutID uint64,
		pegoutUntil int64,
		validatorIdx uint16,
		commitments []byte,
	) (*tlb.Transaction, error)

	SendSigningShare(
		pegoutID uint64,
		pegoutUntil int64,
		validatorIdx uint16,
		signingShares [][]byte,
	) (*tlb.Transaction, error)

	SendSignatures(
		pegoutID uint64,
		pegoutUntil int64,
		validatorIdx uint16,
		signatures [][]byte,
	) (*tlb.Transaction, error)

	SendSigningClaim(
		pegoutID uint64,
		pegoutUntil int64,
		validatorIdx uint16,
		culpritIdx uint16,
	) (*tlb.Transaction, error)

	SendResetPegoutSigning(
		pegoutID uint64,
		validatorIdx uint16,
	) (*tlb.Transaction, error)

	ConnectSigner(signer signer.Signer)
	GetAddr() *address.Address
}

type coordinatorContract struct {
	ton.Contract
	signer            signer.Signer
	tonClient         *tonclient.TonClient
	ctx               context.Context
	ttl               time.Duration
	tonApiCallTimeout int64
}

func New(
	addr *address.Address,
	tonClient *tonclient.TonClient,
	signer signer.Signer,
	ctx context.Context,
	tonApiCallTimeout int64,
) Coordinator {
	ttl := DefaultDGKTTL
	return &coordinatorContract{
		ton.Contract{Addr: addr}, signer, tonClient, ctx, ttl, tonApiCallTimeout,
	}
}
