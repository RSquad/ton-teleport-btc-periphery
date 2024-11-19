package coordinatorcontract

import (
	"context"
	"math/big"
	"time"

	jwv4r2contract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/jw_v4r2_contract"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/xssnick/tonutils-go/address"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
)

const (
	OpCodeDKGStart          = 0xe16b89c0
	CoordinatorRound1       = 0x0000eaea
	CoordinatorRound2       = 0x0000bb50
	CoordinatorRound3       = 0x00008bc6
	PegOutTxSendCommitments = 0x58e40000
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
	packages map[int][][]byte
}
type R2Package struct {
	mask     uint64
	count    uint64
	packages map[int]map[int][][]byte
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
	modeResult, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_standalone_mode")
	if err != nil {
		return false, err
	}

	return modeResult.MustInt(0).Int64() != 0, nil
}

func (c *CoordinatorContract) GetDKG() (DKG, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return DKG{}, err
	}
	dkgResult, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_dkg")
	if err != nil {
		return DKG{}, err
	}

	dkg, err := c.parseDkg(dkgResult.MustCell(0).BeginParse())
	if err != nil {
		return DKG{}, err
	}
	return dkg, nil
}

func (c *CoordinatorContract) parseDkg(dkgCell *cell.Slice) (DKG, error) {
	state, _ := dkgCell.LoadUInt(2)
	dictCell := dkgCell.MustLoadDict(16)
	vsetDict, _ := dictCell.LoadAll()

	vset := make(map[uint64][]byte)

	for _, kv := range vsetDict {
		key, _ := kv.Key.LoadUInt(16)
		value, _ := ValidatorDescrValueParse(kv.Value)
		vset[key] = value
	}

	maxSigners, _ := dkgCell.LoadUInt(16)
	r1PackageParams, _ := PackageParse(dkgCell)

	r1PackageCell := dkgCell.MustLoadDict(16)
	r1PackageDict, _ := r1PackageCell.LoadAll()

	r1Package := make(map[int][][]byte)

	for i, kv := range r1PackageDict {
		//_, err := kv.Key.LoadUInt(32)
		value, _ := packageValueParse(kv.Value)
		r1Package[i] = value
	}

	r2PackageParams, _ := PackageParse(dkgCell)

	r2PackageCell := dkgCell.MustLoadDict(16)
	r2PackageDict, _ := r2PackageCell.LoadAll()

	r2Package := make(map[int]map[int][][]byte)

	for i, kv := range r2PackageDict {
		//_, err := kv.Key.LoadSlice(32)
		value, _ := packageDictionaryValueParse(kv.Value)
		r2Package[i] = value
	}

	cfgHash, _ := dkgCell.LoadSlice(32)
	attempts := dkgCell.MustLoadUInt(8)
	until := dkgCell.MustLoadUInt(32)

	packagesSlice := dkgCell.MustLoadRef()
	validatorsCount := packagesSlice.MustLoadUInt(16)
	validatorsMask := packagesSlice.MustLoadBigUInt(256)

	pubkeyPackage := packagesSlice.MustLoadMaybeRef()
	pubkeyData := PubkeyData{}
	if pubkeyPackage != nil {
		r3pubkeyPackage, _ := writeCellsToBuffer(pubkeyPackage)
		r3InternalKey := packagesSlice.MustLoadSlice(32)
		pubkeyData.pubkeyPackage = r3pubkeyPackage
		pubkeyData.internalKey = r3InternalKey
	}

	return DKG{
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

func (c *CoordinatorContract) GetPrevDKG() (DKG, error) {
	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return DKG{}, err
	}
	dkgResult, err := c.tonClient.API.RunGetMethod(c.ctx, block, c.Address, "get_prev_dkg")
	if err != nil {
		return DKG{}, err
	}

	dkg, err := c.parseDkg(dkgResult.MustCell(0).BeginParse())
	if err != nil {
		return DKG{}, err
	}
	return dkg, nil
}

func (c *CoordinatorContract) sendStartDKG(lifetime int64) error {
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
		//.storeBuffer(signature, 64)
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
		MustStoreUInt(CoordinatorRound1, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			MustStoreBinarySnake(opts.identifier).
			//	.storeBuffer(opts.identifier, 32)
			//	.storeRef(splitBufferToCells(opts.round1Package))
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
		MustStoreUInt(CoordinatorRound2, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			//.storeBuffer(opts.fromIdentifier, 32)
			//	.storeBuffer(opts.toIdentifier, 32)
			//	.storeRef(splitBufferToCells(opts.round2Package))
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
		MustStoreUInt(CoordinatorRound3, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			//.storeBuffer(opts.identifier, 32)
			//  .storeBuffer(opts.internalKeyXY.subarray(1, 65), 64)
			//  .storeRef(splitBufferToCells(opts.pubkeyPackage))
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
		MustStoreUInt(PegOutTxSendCommitments, 32).
		MustStoreUInt(floorTime, 32).
		MustStoreUInt(opts.validatorIdx, 16).
		MustStoreRef(cell.
			BeginCell().
			//.storeBuffer(opts.identifier, 32)
			MustStoreUInt(opts.pegoutId, 64).
			//.storeRef(splitBufferToCells(opts.commitments))
			EndCell(),
		).
		EndCell()
	_, err := c.buildExternalMessage(signBody)
	if err != nil {
		panic(err)
	}
	return nil
}

func (c *CoordinatorContract) sendSigningShare() error { return nil }
func (c *CoordinatorContract) sendSignatures() error   { return nil }
