package coordinator

import (
	"context"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/parseddict"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	tonutils "github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	OpCodeStartDKG                      = 0xe16b89c0
	OpCodeCoordinatorRound1             = 0x0000eaea
	OpCodeCoordinatorRound2             = 0x0000bb50
	OpCodeCoordinatorRound3             = 0x00008bc6
	OpCodeCoordinatorDkgClaim           = 0x0000f387
	OpCodeCoordinatorSendCommitments    = 0x58e40000
	OpCodeCoordinatorSendSigningShare   = 0x706b0000
	OpCodeCoordinatorSendSignature      = 0xd0720000
	OpCodeCoordinatorSigningClaim       = 0x5fcb0000
	OpCodeCoordinatorResetPegoutSigning = 0xe6c20000
)

const DefaultDGKTTL = time.Minute

type Storage struct {
	Initiated           bool
	StandaloneMode      bool
	Id                  uint32
	ConfiguratorAddr    *address.Address
	Enabled             bool
	Dkg                 *DKG
	PrevDkg             *DKG
	UnsignedPegouts     []PegoutRecord
	PegoutTxCode        *cell.Cell
	MinClaimsPercent    uint16
	MinSignersThreshold uint16
	DkgLifetime         uint32
	SigningTimeout      uint32
	NextPegoutIdx       uint64
	TeleportAddr        *address.Address
}

func CallApiWithTimeout[T any](fn func(ctx context.Context) (T, error), parentCtx context.Context, timeout int64, name string) (T, error) {
	startTs := time.Now()
	apiCtx, cancelFn := context.WithTimeout(parentCtx, time.Duration(timeout)*time.Second)
	defer cancelFn()
	res, err := fn(apiCtx)
	endTs := time.Now()
	duration := endTs.Unix() - startTs.Unix()

	logger.Log.Debug().Msgf("Ton API call: '%s', total time {%d}s", name, duration)

	return res, err
}

func (c *coordinatorContract) GetDkg(block *tonutils.BlockIDExt) (*DKG, error) {
	if block == nil {
		var err error

		block, err = CallApiWithTimeout(
			func(apiCallCtx context.Context) (*tonutils.BlockIDExt, error) {
				return c.tonClient.API.CurrentMasterchainInfo(apiCallCtx)
			},
			c.ctx,
			c.tonApiCallTimeout,
			"CurrentMasterchainInfo",
		)
		if err != nil {
			return nil, err
		}
	}

	result, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tonutils.ExecutionResult, error) {
			return c.tonClient.API.RunGetMethod(apiCallCtx, block, c.Addr, "get_dkg")
		},
		c.ctx,
		c.tonApiCallTimeout,
		"get_dkg",
	)
	if err != nil {
		return nil, err
	}

	dkgNotExists, err := result.IsNil(0)
	if err != nil {
		return nil, err
	}

	if dkgNotExists {
		return nil, nil
	}

	dkg, err := parseDGKSlice(result.MustCell(0).BeginParse())
	if err != nil {
		return nil, err
	}

	return dkg, nil
}

func (c *coordinatorContract) GetPrevDKG() (*DKG, error) {
	block, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tonutils.BlockIDExt, error) {
			return c.tonClient.API.CurrentMasterchainInfo(apiCallCtx)
		},
		c.ctx,
		c.tonApiCallTimeout,
		"CurrentMasterchainInfo",
	)
	if err != nil {
		return nil, err
	}

	result, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tonutils.ExecutionResult, error) {
			return c.tonClient.API.RunGetMethod(apiCallCtx, block, c.Addr, "get_prev_dkg")
		},
		c.ctx,
		c.tonApiCallTimeout,
		"get_prev_dkg",
	)
	if err != nil {
		return nil, err
	}

	dkgNotExists, err := result.IsNil(0)
	if err != nil {
		return nil, err
	}

	if dkgNotExists {
		return nil, nil
	}

	return parseDGKSlice(result.MustCell(0).BeginParse())
}

func (c *coordinatorContract) GetUnsignedPegouts() ([]PegoutRecord, error) {
	block, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tonutils.BlockIDExt, error) {
			return c.tonClient.API.CurrentMasterchainInfo(apiCallCtx)
		},
		c.ctx,
		c.tonApiCallTimeout,
		"CurrentMasterchainInfo",
	)
	if err != nil {
		return nil, err
	}

	result, err := CallApiWithTimeout(
		func(apiCallCtx context.Context) (*tonutils.ExecutionResult, error) {
			return c.tonClient.API.RunGetMethod(apiCallCtx, block, c.Addr, "get_pegout_records")
		},
		c.ctx,
		c.tonApiCallTimeout,
		"get_pegout_records",
	)
	if err != nil {
		return nil, err
	}

	isNullCell, err := result.IsNil(0)
	if err != nil {
		return nil, err
	}

	if isNullCell {
		return nil, nil
	}

	cell, err := result.Cell(0)
	if err != nil {
		return nil, err
	}
	return parseUnsignedPegouts(cell)
}

func (c *coordinatorContract) GetStorage(block *tonutils.BlockIDExt) (Storage, error) {
	if block == nil {
		var err error
		block, err = c.tonClient.API.CurrentMasterchainInfo(c.ctx)
		if err != nil {
			return Storage{}, err
		}
	}
	acc, err := c.tonClient.FetchAcc(c.Addr, block)
	if err != nil {
		return Storage{}, err
	}
	storage := acc.Data.BeginParse()

	initiated := storage.MustLoadBoolBit()
	standaloneMode := storage.MustLoadBoolBit()
	id := uint32(storage.MustLoadUInt(32))
	configuratorAddr := storage.MustLoadAddr()
	enabled := storage.MustLoadBoolBit()
	dkgSlice, err := storage.LoadMaybeRef()
	if err != nil {
		return Storage{}, err
	}
	var dkg *DKG
	if dkgSlice == nil {
		dkg = &DKG{}
	} else {
		dkg, err = parseDGKSlice(dkgSlice)
	}
	if err != nil {
		return Storage{}, err
	}
	var prevDkg *DKG
	prevDkgSlice, err := storage.LoadMaybeRef()
	if err != nil {
		return Storage{}, err
	}
	if prevDkgSlice == nil {
		prevDkg = &DKG{}
	} else {
		prevDkg, err = parseDGKSlice(prevDkgSlice)
	}
	if err != nil {
		return Storage{}, err
	}

	var unsignedPegouts []PegoutRecord
	unsignedPegoutsSlice, err := storage.LoadMaybeRef()
	if err != nil {
		return Storage{}, err
	}
	if unsignedPegoutsSlice == nil {
		unsignedPegouts = []PegoutRecord{}
	} else {
		unsignedPegouts, err = parseUnsignedPegouts(unsignedPegoutsSlice.MustToCell())
	}
	if err != nil {
		return Storage{}, err
	}
	pegoutTxCode := storage.MustLoadRef().MustToCell()
	minClaimsPercent := uint16(storage.MustLoadUInt(16))
	minSignersThreshold := uint16(storage.MustLoadUInt(16))
	dkgLifetime := uint32(storage.MustLoadUInt(32))
	signingTimeout := uint32(storage.MustLoadUInt(32))
	nextPegoutIdx := storage.MustLoadUInt(64)
	teleportAddr := storage.MustLoadAddr()

	return Storage{
		Initiated:           initiated,
		StandaloneMode:      standaloneMode,
		Id:                  id,
		ConfiguratorAddr:    configuratorAddr,
		Enabled:             enabled,
		Dkg:                 dkg,
		PrevDkg:             prevDkg,
		UnsignedPegouts:     unsignedPegouts,
		PegoutTxCode:        pegoutTxCode,
		MinClaimsPercent:    minClaimsPercent,
		MinSignersThreshold: minSignersThreshold,
		DkgLifetime:         dkgLifetime,
		SigningTimeout:      signingTimeout,
		NextPegoutIdx:       nextPegoutIdx,
		TeleportAddr:        teleportAddr,
	}, nil
}

func parseUnsignedPegouts(pegoutsCell *cell.Cell) ([]PegoutRecord, error) {
	dict, err := pegoutsCell.BeginParse().ToDict(64)
	if err != nil {
		return nil, err
	}
	entries, err := dict.LoadAll()
	if err != nil {
		return nil, err
	}

	pegouts := make([]PegoutRecord, 0, len(entries))
	for _, kv := range entries {

		ID := kv.Key.MustLoadUInt(64)
		value := kv.Value.MustLoadRef()

		MaxSigners := uint16(value.MustLoadUInt(16))
		ExpiredAt := time.Unix(int64(value.MustLoadUInt(32)), 0)
		SigningMask := value.MustLoadBigUInt(256)

		CommitmentsMask := value.MustLoadSlice(256)
		value.MustLoadUInt(16)
		commitmentsDict := value.MustLoadDict(16)
		commitmentsPtr, err := parseddict.ParseDict(
			commitmentsDict,
			parseddict.ParseKeyUI16,
			readBuffer,
		)
		if err != nil {
			return nil, err
		}

		Commitments := *commitmentsPtr
		SigningSharesMask := value.MustLoadSlice(256)
		value.MustLoadUInt(16)
		signingSharesDict := value.MustLoadDict(16)
		signingSharesPtr, err := parseddict.ParseDict(
			signingSharesDict,
			parseddict.ParseKeyUI16,
			loadSharesMap,
		)
		if err != nil {
			return nil, err
		}

		SigningShares := *signingSharesPtr

		sigSlice := value.MustLoadRef()
		Signatures := PegoutSignatures{
			Mask:  sigSlice.MustLoadBigUInt(256),
			Count: uint16(sigSlice.MustLoadUInt(16)),
			Hash:  sigSlice.MustLoadSlice(256),
		}

		refSlice := value.MustLoadRef()
		claimsSlice := refSlice.MustLoadRef()
		ClaimsMask := claimsSlice.MustLoadBigUInt(256)
		ClaimsCount := uint16(claimsSlice.MustLoadUInt(16))
		claimsCountersDict := claimsSlice.MustLoadDict(16)
		claimsCountersPtr, err := parseddict.ParseDict(
			claimsCountersDict,
			parseddict.ParseKeyUI16,
			loadUI16Map,
		)
		if err != nil {
			return nil, err
		}

		ClaimsCounters := *claimsCountersPtr

		InternalKey := refSlice.MustLoadSlice(256)
		PegoutAddress := refSlice.MustLoadAddr()
		IsAutopegout := refSlice.MustLoadInt(1) == -1

		pegouts = append(pegouts, PegoutRecord{
			ID,
			PegoutAddress,
			InternalKey,
			IsAutopegout,
			Commitments,
			CommitmentsMask,
			SigningShares,
			SigningSharesMask,
			Signatures,
			ClaimsMask,
			ClaimsCount,
			ClaimsCounters,
			MaxSigners,
			ExpiredAt,
			SigningMask,
		})
	}
	return pegouts, nil
}

func parseDGKSlice(dkgSlice *cell.Slice) (*DKG, error) {
	// State
	state := DKGState(dkgSlice.MustLoadUInt(2))

	// VSet
	vSetDictCell := dkgSlice.MustLoadDict(16)
	vSet, err := NewVSet(vSetDictCell)
	if err != nil {
		return nil, err
	}

	// Max signers
	maxSigners := uint16(dkgSlice.MustLoadUInt(16))

	// VSet mask
	vSetMask := dkgSlice.MustLoadBigUInt(256)

	// R1
	r1State, err := loadRoundMaskAndCount(dkgSlice)
	if err != nil {
		return nil, err
	}
	r1PkgDictCell := dkgSlice.MustLoadDict(16)
	r1, err := NewDKGR1(r1PkgDictCell, r1State)
	if err != nil {
		return nil, err
	}

	// R2
	r2State, err := loadRoundMaskAndCount(dkgSlice)
	if err != nil {
		return nil, err
	}
	r2PkgDictCell := dkgSlice.MustLoadDict(16)
	r2, err := NewDKGR2(r2PkgDictCell, r2State)
	if err != nil {
		return nil, err
	}

	// Attempts
	attempts := dkgSlice.MustLoadUInt(8)

	// DKG lifetime (timestamp)
	untilUnix := dkgSlice.MustLoadUInt(32)
	until := time.Unix(int64(untilUnix), 0)

	dkgNextRef := dkgSlice.MustLoadRef()

	// VSet config hash
	cfgHash := dkgNextRef.MustLoadSlice(256)

	// R3
	r3, err := LoadDKGR3(dkgNextRef)
	if err != nil {
		return nil, err
	}

	// Claims
	claims, err := LoadDKGClaims(dkgNextRef)
	if err != nil {
		return nil, err
	}

	// Session keys
	sessionKeys, err := LoadSessionKeys(dkgNextRef)
	if err != nil {
		return nil, err
	}

	return &DKG{
		State:       state,
		VSet:        vSet,
		MaxSigners:  maxSigners,
		VSetMask:    vSetMask,
		SessionKeys: sessionKeys,
		R1:          r1,
		R2:          r2,
		R3:          r3,
		Claims:      claims,
		CfgHash:     cfgHash,
		Attempts:    attempts,
		Until:       until,
	}, nil
}

func loadRoundMaskAndCount(dkgSlice *cell.Slice) (*DKGRoundState, error) {
	mask, err := dkgSlice.LoadBigUInt(256)
	if err != nil {
		return nil, err
	}
	count, err := dkgSlice.LoadUInt(16)
	if err != nil {
		return nil, err
	}

	return &DKGRoundState{mask, count}, nil
}

// Helpers

func readBuffer(value *cell.Slice) ([]byte, error) {
	return utils.WriteSlicesToBuffer(value.MustLoadRef()), nil
}

func loadSharesMap(value *cell.Slice) (map[uint16][]byte, error) {
	dict, err := value.MustLoadRef().ToDict(64)
	if err != nil {
		return nil, err
	}
	sharesMap, err := parseddict.ParseDict(
		dict,
		parseddict.ParseKeyUI16,
		func(s *cell.Slice) ([]byte, error) {
			return utils.WriteSlicesToBuffer(s), nil
		},
	)
	if err != nil {
		return nil, err
	}

	return *sharesMap, err
}

func loadUI16Map(s *cell.Slice) (uint16, error) {
	return uint16(s.MustLoadUInt(16)), nil
}
