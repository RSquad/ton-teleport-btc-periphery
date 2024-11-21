package coordinatorcontract

import (
	"context"
	"math/big"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/dict"

	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
)

const (
	OpCodeDKGStart                = 0xe16b89c0
	OpCodeCoordinatorRound1       = 0x0000eaea
	OpCodeCoordinatorRound2       = 0x0000bb50
	OpCodeCoordinatorRound3       = 0x00008bc6
	OpCodePegOutTxSendCommitments = 0x58e40000
)

type CoordinatorContract struct {
	Address   *address.Address
	tonClient *tonclient.TonClient
	sender    *jwv4r2contract.JWV4R2Contract
	ctx       context.Context
}
type R1Package struct {
	mask     uint64
	count    uint64
	packages map[dict.R1PackageKey]dict.R1PackageValue
}
type R2Package struct {
	mask     uint64
	count    uint64
	packages map[dict.R2PackageKey]dict.R2PackageValue
}
type R3Package struct {
	mask       *big.Int
	count      uint64
	pubkeyData PubkeyData
}
type PubkeyData struct {
	pubkeyPackage [][]byte
	internalKey   []byte
}
type DKG struct {
	state      uint64
	maxSigners uint64
	vset       map[uint64][]byte
	r1Packages R1Package
	r2Packages R2Package
	r3Package  R3Package
	cfgHash    []byte
	attempts   uint64
	until      uint64
}
type DkgPackage struct {
	params   DkgPackageParams
	packages uint64
}
type R1Options struct {
	lifetime      int64
	validatorIdx  uint64
	identifier    []byte
	round1Package []byte
}
type R2Options struct {
	lifetime       int64
	validatorIdx   uint64
	fromIdentifier []byte
	toIdentifier   []byte
	round2Package  []byte
}
type PubKeyOptions struct {
	lifetime      int64
	validatorIdx  uint64
	pubkeyPackage []byte
	internalKey   []byte
	identifier    []byte
}
type CommitmentsOptions struct {
	lifetime     int64
	validatorIdx uint64
	identifier   []byte
	commitments  []byte
	pegoutId     uint64
}

func NewCoordinatorContract(
	address *address.Address,
	tonClient *tonclient.TonClient,
	sender *jwv4r2contract.JWV4R2Contract,
	ctx context.Context,
) (*CoordinatorContract, error) {
	return &CoordinatorContract{
		Address:   address,
		tonClient: tonClient,
		sender:    sender,
		ctx:       ctx,
	}, nil
}

func (c *CoordinatorContract) GetStandaloneMode() (bool, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return false, err
	}
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_standalone_mode")
	if err != nil {
		return false, err
	}

	return result.MustInt(0).Int64() != 0, nil
}

func (c *CoordinatorContract) GetDKG() (*DKG, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_dkg")
	if err != nil {
		return nil, err
	}

	dkg, err := c.parseDkg(result.MustCell(0).BeginParse())
	if err != nil {
		return nil, err
	}
	return dkg, nil
}

func (c *CoordinatorContract) parseDkg(dkgCell *cell.Slice) (*DKG, error) {
	state, _ := dkgCell.LoadUInt(2)

	dictionary := dkgCell.MustLoadDict(16)

	vsetDict := dict.VSetDict{}
	vset := vsetDict.NewDict(dictionary).Get()

	maxSigners, _ := dkgCell.LoadUInt(16)

	r1PackageParams, _ := packageParse(dkgCell)
	r1PackageDict := dict.R1PackageDict{}
	r1Package := r1PackageDict.NewDict(dkgCell.MustLoadDict(256)).Get()

	r2PackageParams, _ := packageParse(dkgCell)
	r2PackageDict := dict.R2PackageDict{}
	r2Package := r2PackageDict.NewDict(dkgCell.MustLoadDict(256)).Get()

	cfgHash, _ := dkgCell.LoadSlice(256)
	attempts := dkgCell.MustLoadUInt(8)
	until := dkgCell.MustLoadUInt(32)

	packagesSlice := dkgCell.MustLoadRef()
	validatorsCount := packagesSlice.MustLoadUInt(16)
	validatorsMask := packagesSlice.MustLoadBigUInt(256)

	pubkeyPackage := packagesSlice.MustLoadMaybeRef()
	pubkeyData := PubkeyData{}
	if pubkeyPackage != nil {
		r3pubkeyPackage := dict.WriteCellsToBuffer(pubkeyPackage)
		r3InternalKey := packagesSlice.MustLoadSlice(256)
		pubkeyData.pubkeyPackage = r3pubkeyPackage
		pubkeyData.internalKey = r3InternalKey
	}

	return &DKG{
		maxSigners: maxSigners,
		state:      state,
		vset:       vset,
		r1Packages: R1Package{
			count:    r1PackageParams.count,
			mask:     r1PackageParams.mask,
			packages: r1Package,
		},
		r2Packages: R2Package{
			count:    r2PackageParams.count,
			mask:     r2PackageParams.mask,
			packages: r2Package,
		},
		r3Package: R3Package{
			mask:       validatorsMask,
			count:      validatorsCount,
			pubkeyData: pubkeyData,
		},
		cfgHash:  cfgHash,
		attempts: attempts,
		until:    until,
	}, nil
}

func (c *CoordinatorContract) GetPrevDKG() (*DKG, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_prev_dkg")
	if err != nil {
		return nil, err
	}

	dkg, err := c.parseDkg(result.MustCell(0).BeginParse())
	if err != nil {
		return nil, err
	}
	return dkg, nil
}

func (c *CoordinatorContract) SendStartDKG(lifetime int64) error {
	floorTime := uint64(time.Now().Unix() + lifetime)
	signBody := cell.
		BeginCell().
		MustStoreUInt(OpCodeDKGStart, 32).
		MustStoreUInt(floorTime, 32).
		EndCell()
	_, err := c.buildExternalMessage(signBody)
	if err != nil {
		panic(err)
	}

	return nil
}

func (c *CoordinatorContract) buildExternalMessage(signBody *cell.Cell) (*tlb.ExternalMessage, error) {
	body := cell.
		BeginCell().
		MustStoreRef(signBody).
		EndCell()

	msg := &tlb.ExternalMessage{
		DstAddr: c.Address,
		Body:    body,
	}
	return msg, nil
}

func (c *CoordinatorContract) sendRound1(opts R1Options) error {
	if len(opts.identifier) != 32 {
		panic("identifier must be 32 bytes length")
	}
	floorTime := uint64(time.Now().Unix() + opts.lifetime)
	signBody := cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound1, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			MustStoreBinarySnake(opts.identifier).
			MustStoreSlice(opts.identifier, 32).
			EndCell()).
		EndCell()
	_, err := c.buildExternalMessage(signBody)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *CoordinatorContract) sendRound2(opts R2Options) error {
	if len(opts.fromIdentifier) != 32 || len(opts.toIdentifier) != 32 {
		panic("identifier must be 32 bytes length")
	}
	floorTime := uint64(time.Now().Unix() + opts.lifetime)
	signBody := cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound2, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			MustStoreSlice(opts.fromIdentifier, 32).
			MustStoreSlice(opts.toIdentifier, 32).
			EndCell(),
		).
		EndCell()
	_, err := c.buildExternalMessage(signBody)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *CoordinatorContract) sendPubkeyPackage(opts PubKeyOptions) error {
	if len(opts.internalKey) != 32 {
		panic("Internal key must be 65 bytes and has prefix 0x04")
	}
	floorTime := uint64(time.Now().Unix() + opts.lifetime)
	signBody := cell.BeginCell().
		MustStoreUInt(OpCodeCoordinatorRound3, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			MustStoreSlice(opts.identifier, 32).
			MustStoreSlice(opts.internalKey, 64).
			EndCell(),
		).
		EndCell()
	_, err := c.buildExternalMessage(signBody)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *CoordinatorContract) sendCommitments(opts CommitmentsOptions) error {
	if len(opts.identifier) != 32 {
		panic("identifier must be 32 bytes length")
	}
	floorTime := uint64(time.Now().Unix() + opts.lifetime)
	signBody := cell.BeginCell().
		MustStoreUInt(OpCodePegOutTxSendCommitments, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			MustStoreSlice(opts.identifier, 32).
			MustStoreUInt(opts.pegoutId, 64).
			EndCell(),
		).
		EndCell()
	_, err := c.buildExternalMessage(signBody)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *CoordinatorContract) GetSigningShares(pegoutTxId *big.Int) (map[dict.SigningSharesKey]dict.SigningSharesValue, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_signature_shares", pegoutTxId)
	if err != nil {
		return nil, err
	}
	slice := result.MustCell(0).BeginParse()

	signingSharesDict := dict.SigningSharesDict{}
	shares := signingSharesDict.NewDict(slice.MustLoadDict(256)).Get()

	return shares, nil
}

func (c *CoordinatorContract) GetUnsignedPegouts() (map[dict.UnsignedPegoutsKey]dict.UnsignedPegoutsValue, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return nil, err
	}
	result, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_pegout_records")
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	slice := result.MustCell(0).BeginParse()
	unsignedPegoutsDict := dict.UnsignedPegoutsDict{}
	pegouts := unsignedPegoutsDict.NewDict(slice.MustLoadDict(16)).Get()

	return pegouts, nil
}

func (c *CoordinatorContract) sendSigningShare() error { return nil }
func (c *CoordinatorContract) sendSignatures() error   { return nil }
