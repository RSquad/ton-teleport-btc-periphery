package metrics

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sync"
	"time"

	"slices"

	bu "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/bitcoinutils"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherPegouts struct {
	tonClient           *tonclient.TonClient
	bitcoinClient       *bitcoin.Client
	coordinatorContract coordinator.Coordinator
	db                  *sql.DB
	period              int64
	pegoutTempData      PegoutTempData
}

type PegoutTempData struct {
	lastAutopegoutDate time.Time
	expirations        map[string]Expirations
	unsignedPegouts    *Cache[map[uint64]coordinator.PegoutRecord]
	signedPegouts      *Cache[map[string]SignedPegout]
	curDkg             *Cache[coordinator.DKG]
	prevDkg            *Cache[coordinator.DKG]
}

type Expirations struct {
	actual       time.Time
	previous     time.Time
	restartCount int
}

type InternalKeys struct {
	dkg     []byte
	prevDkg []byte
}

var EMPTY_PEGOUT coordinator.PegoutRecord = coordinator.PegoutRecord{
	ID:                      0,
	PegoutAddress:           &address.Address{},
	InternalKey:             []byte{},
	IsAutopegout:            false,
	Commitments:             map[uint16][]byte{},
	CommitmentsMaskAccepted: big.NewInt(0),
	CommitmentsMaskOther:    big.NewInt(0),
	SigningShares:           map[uint16]map[uint16][]byte{},
	SigningSharesMask:       []byte{},
	Signatures: coordinator.PegoutSignatures{
		Mask:  big.NewInt(0),
		Count: 0,
		Hash:  []byte{},
	},
	ClaimsMask:     big.NewInt(0),
	ClaimsCount:    0,
	ClaimsCounters: map[uint16]uint16{},
	MaxSigners:     0,
	ExpiredAt:      time.Time{},
	SigningMask:    big.NewInt(0),
}

var EMPTY_DKG coordinator.DKG = coordinator.DKG{
	State:       coordinator.DKGStateFinished,
	VSet:        coordinator.VSet{},
	MaxSigners:  0,
	VSetMask:    big.NewInt(0),
	SessionKeys: &coordinator.SessionKeys{},
	R1: &coordinator.DKGR1{
		Count:    0,
		Mask:     big.NewInt(0),
		Packages: map[uint16][]byte{},
	},
	R2: &coordinator.DKGR2{
		Count:    0,
		Mask:     big.NewInt(0),
		Packages: map[uint16][]byte{},
	},
	R3: &coordinator.DKGR3{
		Count: 0,
		Mask:  big.NewInt(0),
		Data: &coordinator.PubkeyData{
			PubkeyPackage: []byte{},
			InternalKey:   []byte{},
		},
	},
	Claims: &coordinator.DKGClaims{
		Mask:     big.NewInt(0),
		Count:    0,
		Counters: map[uint16]uint16{},
	},
	CfgHash:  []byte{},
	Attempts: 0,
	Until:    time.Time{},
}

func getMapKeysUint64(data map[uint64]coordinator.PegoutRecord) []uint64 {
	keys := make([]uint64, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	slices.Sort(keys)
	return keys
}

func NewFetcherPegouts(
	tonClient *tonclient.TonClient,
	bitcoinClient *bitcoin.Client,
	coordinatorContract coordinator.Coordinator,
	db *sql.DB,
	period int64,
) *FetcherPegouts {
	return &FetcherPegouts{
		tonClient:           tonClient,
		bitcoinClient:       bitcoinClient,
		db:                  db,
		coordinatorContract: coordinatorContract,
		period:              period,
		pegoutTempData: PegoutTempData{
			lastAutopegoutDate: time.Time{},
			expirations:        make(map[string]Expirations),
			unsignedPegouts:    NewCache[map[uint64]coordinator.PegoutRecord](),
			signedPegouts:      NewCache[map[string]SignedPegout](),
			curDkg:             NewCache[coordinator.DKG](),
			prevDkg:            NewCache[coordinator.DKG](),
		},
	}
}

func (f *FetcherPegouts) setDelayedMetric(pegouts map[uint64]coordinator.PegoutRecord) {

	if len(pegouts) == 0 {
		unsignedPegoutDelayed.WithLabelValues(utils.AddrToRawString(&address.Address{})).Set(0)
		return
	}
	now := time.Now()
	for _, pegout := range pegouts {
		if now.After(pegout.ExpiredAt.Add(PEGOUT_MAX_DELAY)) {
			unsignedPegoutDelayed.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(1)
		} else {
			unsignedPegoutDelayed.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(0)
		}
	}
}

func (f *FetcherPegouts) getPegoutsData() (map[uint64]coordinator.PegoutRecord, error) {
	rows, err := f.db.Query(
		`SELECT payload::json
		FROM metrics_data WHERE type_id = 5
		ORDER BY id DESC
		LIMIT 1
	`)
	if err != nil {
		return make(map[uint64]coordinator.PegoutRecord), err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(
			&data,
		)
		if err != nil {
			return make(map[uint64]coordinator.PegoutRecord), err
		}
	}

	var coordinatorData ContractCoordinatorData
	err = json.Unmarshal([]byte(data), &coordinatorData)
	if err != nil {
		return make(map[uint64]coordinator.PegoutRecord), err
	}

	pegouts := make(map[uint64]coordinator.PegoutRecord)
	for _, pegout := range coordinatorData.UnsignedPegouts {
		pegouts[pegout.ID] = pegout
		pegoutAddr := utils.AddrToRawString(pegout.PegoutAddress)
		if _, exists := f.pegoutTempData.expirations[pegoutAddr]; !exists {
			f.pegoutTempData.expirations[pegoutAddr] = Expirations{
				actual:       pegout.ExpiredAt,
				previous:     pegout.ExpiredAt,
				restartCount: 0,
			}
		}
	}

	return pegouts, nil
}

func (f *FetcherPegouts) getSignedPegouts() (map[string]SignedPegout, error) {
	rows, err := f.db.Query(
		`SELECT
			tt.created_at,
			p.addr AS pegout_addr,
			p.bitcoin_tx_id AS bitcoin_tx_id
		FROM burns AS b
		JOIN ton_txes AS tt ON tt.id = b.ton_tx_burn
		JOIN pegouts AS p ON p.id = b.pegout_burn
		WHERE
			p.status = 'SIGNED'
		AND
			created_at > NOW() - INTERVAL '1 day'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return map[string]SignedPegout{}, err
	}

	defer rows.Close()

	var pegouts = make(map[string]SignedPegout)
	for rows.Next() {
		var pegout SignedPegout
		err = rows.Scan(&pegout.createdAt, &pegout.pegoutAddr, &pegout.bitcoinTxId)
		if err != nil {
			return map[string]SignedPegout{}, err
		}
		pegouts[pegout.pegoutAddr] = pegout
	}
	return pegouts, nil
}

func (f *FetcherPegouts) getDkg(typeId int) (coordinator.DKG, error) {
	query := `
    SELECT payload::json
    FROM metrics_data 
    WHERE type_id = $1
    ORDER BY id DESC
    LIMIT 1
`
	rows, err := f.db.Query(query, typeId)
	if err != nil {
		return EMPTY_DKG, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(
			&data,
		)
		if err != nil {
			return EMPTY_DKG, err
		}
	}

	var dkgData coordinator.DKG
	err = json.Unmarshal([]byte(data), &dkgData)
	if err != nil {
		return EMPTY_DKG, err
	}

	return dkgData, nil
}

func (f *FetcherPegouts) setMinSignersMetric(dkgMaxSiners uint16) {
	if dkgMaxSiners == 0 {
		pegoutMinSigners.Set(0)
		return
	}
	minSigners := 2 * 3 / dkgMaxSiners
	pegoutMinSigners.Set(float64(minSigners))
}

func (f *FetcherPegouts) setBitcoinTxExistsMetric(pegouts map[string]SignedPegout) {
	if len(pegouts) == 0 {
		unprocessedPegout.WithLabelValues(utils.AddrToRawString(&address.Address{}), "").Set(0)
		return
	}
	for _, pegout := range pegouts {
		txExists, _, _ := bu.BitcoinTxExists(f.bitcoinClient, pegout.bitcoinTxId)
		if !txExists {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(1)
		} else {
			unprocessedPegout.WithLabelValues(pegout.pegoutAddr, pegout.bitcoinTxId).Set(0)
		}

	}
}

func (f *FetcherPegouts) setAutopegoutDelayedMetric(pegouts map[uint64]coordinator.PegoutRecord) {
	autopegoutDelayed.Set(0)
	now := time.Now()
	for _, pegout := range pegouts {
		if pegout.IsAutopegout {
			f.pegoutTempData.lastAutopegoutDate = now
		}
	}

	if now.After(f.pegoutTempData.lastAutopegoutDate.Add(AUTOPEGOUT_WARN_DELAY)) {
		autopegoutDelayed.Set(1)
	}
	if now.After(f.pegoutTempData.lastAutopegoutDate.Add(AUTOPEGOUT_CRIT_DELAY)) {
		autopegoutDelayed.Set(2)
	}
	if now.After(f.pegoutTempData.lastAutopegoutDate.Add(AUTOPEGOUT_PANIC_DELAY)) {
		autopegoutDelayed.Set(3)
	}
}

func (f *FetcherPegouts) setWrongInternalKeyMetric(
	pegouts map[uint64]coordinator.PegoutRecord,
	internalKeys InternalKeys,
) {
	for _, pegout := range pegouts {
		if !bytes.Equal(pegout.InternalKey, internalKeys.dkg) ||
			!bytes.Equal(pegout.InternalKey, internalKeys.prevDkg) {
			wrongInternalKey.WithLabelValues(
				base64.StdEncoding.EncodeToString(pegout.InternalKey),
				utils.AddrToRawString(pegout.PegoutAddress),
			).Set(1)
		} else {
			wrongInternalKey.WithLabelValues(
				base64.StdEncoding.EncodeToString(pegout.InternalKey),
				utils.AddrToRawString(pegout.PegoutAddress),
			).Set(0)
		}
	}
}

func (f *FetcherPegouts) setPegoutMaxSignersMetric(pegouts map[uint64]coordinator.PegoutRecord) {
	for _, pegout := range pegouts {
		pegoutMaxSigners.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(float64(pegout.MaxSigners))
	}
}

func (f *FetcherPegouts) setSigningRestartMetric(pegouts map[uint64]coordinator.PegoutRecord) {
	for _, pegout := range pegouts {
		pegoutAddr := utils.AddrToRawString(pegout.PegoutAddress)
		expirations := f.pegoutTempData.expirations[pegoutAddr]
		if expirations.actual.After(expirations.previous) {
			pegoutSigningRestart.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(1)
			f.updateRestartCount(pegoutAddr)
		} else {
			pegoutSigningRestart.WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).Set(0)
		}
	}
}

func (f *FetcherPegouts) updateRestartCount(pegoutAddr string) {
	expirations := f.pegoutTempData.expirations[pegoutAddr]
	expirations.restartCount = expirations.restartCount + 1
	f.pegoutTempData.expirations[pegoutAddr] = expirations
}

func (f *FetcherPegouts) setRestartSigningPegoutCountMetric(expirations map[string]Expirations) {
	for addr, expirations := range expirations {
		pegoutRestartCount.WithLabelValues(addr).Set(float64(expirations.restartCount))
	}
}

func (f *FetcherPegouts) setCulpritMetrics(pegout coordinator.PegoutRecord) {
	if pegout.ClaimsCount > 0 {
		if popcnt(pegout.ClaimsMask) < int(pegout.MaxSigners) {
			pegoutSigningCulpritNotGetThreshold.
				WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).
				Set(1)
		} else {
			pegoutSigningCulpritGotThreshold.
				WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).
				Set(1)
		}
	} else {
		pegoutSigningCulpritNotGetThreshold.
			WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).
			Set(0)

		pegoutSigningCulpritGotThreshold.
			WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).
			Set(0)
	}
}

func (f *FetcherPegouts) setSigningMaskMetrics(pegout coordinator.PegoutRecord) {
	signingMaskCount := popcnt(pegout.SigningMask)

	pegoutSigningMaskValidatorsCount.
		WithLabelValues(utils.AddrToRawString(pegout.PegoutAddress)).
		Set(float64(signingMaskCount))
}

func (f *FetcherPegouts) Fetch() {
	var unsignedPegouts map[uint64]coordinator.PegoutRecord
	var err error
	cache, ok := f.pegoutTempData.unsignedPegouts.Get("UnsignedPegouts")

	if ok {
		unsignedPegouts = cache
	} else {
		unsignedPegouts, err = f.getPegoutsData()
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherPegouts").
				Msg("fetch failed")
		}

		f.pegoutTempData.unsignedPegouts.Set("UnsignedPegouts", unsignedPegouts, time.Duration(f.period)*time.Second)
	}

	if unsignedPegouts == nil {
		logger.Log.Debug().Msg("FetcherPegouts: Contract returns unsignedPegouts is null")
	}

	unsignedPegoutsLen.WithLabelValues("Unsigned pegouts length").Set(float64(len(unsignedPegouts)))

	f.setDelayedMetric(unsignedPegouts)
	f.setAutopegoutDelayedMetric(unsignedPegouts)
	f.setPegoutMaxSignersMetric(unsignedPegouts)
	f.setSigningRestartMetric(unsignedPegouts)
	f.setRestartSigningPegoutCountMetric(f.pegoutTempData.expirations)
	keys := getMapKeysUint64(unsignedPegouts)
	if len(keys) == 0 {
		logger.Log.Debug().Msg("FetcherPegouts: Contract returns unsignedPegouts is empty")
		f.setSigningMaskMetrics(EMPTY_PEGOUT)
		f.setCulpritMetrics(EMPTY_PEGOUT)
	} else {
		f.setSigningMaskMetrics(unsignedPegouts[keys[0]])
		f.setCulpritMetrics(unsignedPegouts[keys[0]])
	}
	var signedPegouts = make(map[string]SignedPegout)
	signedCache, ok := f.pegoutTempData.signedPegouts.Get("SignedPegouts")

	if ok {
		signedPegouts = signedCache
	} else {
		signedPegouts, err = f.getSignedPegouts()
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherPegouts").
				Msg("fetch failed")
		}
		f.pegoutTempData.signedPegouts.Set("SignedPegouts", signedPegouts, time.Duration(f.period)*time.Second)
	}
	f.setBitcoinTxExistsMetric(signedPegouts)

	var curDkg coordinator.DKG
	var prevDkg coordinator.DKG
	curDkgCache, ok := f.pegoutTempData.curDkg.Get("CurDkg")
	if ok {
		curDkg = curDkgCache
	} else {
		curDkg, err = f.getDkg(0)
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherPegouts").
				Msg("fetch failed")
		}
		f.pegoutTempData.curDkg.Set("CurDkg", curDkg, time.Duration(f.period)*time.Second)
	}
	f.setMinSignersMetric(curDkg.MaxSigners)
	prevDkgCache, ok := f.pegoutTempData.prevDkg.Get("PrevDkg")
	if ok {
		prevDkg = prevDkgCache
	} else {
		prevDkg, err = f.getDkg(1)
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherPegouts").
				Msg("fetch failed")
		}
		f.pegoutTempData.prevDkg.Set("PrevDkg", prevDkg, time.Duration(f.period)*time.Second)
	}

	var curKey []byte = []byte{}
	var prevKey []byte = []byte{}
	if curDkg.R3.Data != nil {
		curKey = curDkg.R3.Data.InternalKey
	}
	if prevDkg.R3.Data != nil {
		prevKey = prevDkg.R3.Data.InternalKey
	}

	internalKeys := InternalKeys{
		dkg:     curKey,
		prevDkg: prevKey,
	}

	f.setWrongInternalKeyMetric(unsignedPegouts, internalKeys)
	f.pegoutTempData.unsignedPegouts.DeleteExpired()
	f.pegoutTempData.signedPegouts.DeleteExpired()
	f.pegoutTempData.unsignedPegouts.DeleteExpired()
	f.pegoutTempData.signedPegouts.DeleteExpired()
	f.pegoutTempData.curDkg.DeleteExpired()
	f.pegoutTempData.prevDkg.DeleteExpired()
}

func (fetcher *FetcherPegouts) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherPegouts: stopped")
	logger.DefaultLogStartWork("FetcherPegouts: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Pegouts Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}
