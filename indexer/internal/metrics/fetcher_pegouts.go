package metrics

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

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
	internalKeys       *Cache[InternalKeys]
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

func popcnt(n *big.Int) int {
	count := 0
	zero := big.NewInt(0)
	one := big.NewInt(1)
	temp := new(big.Int)

	for n.Cmp(zero) != 0 {
		count++
		temp.Sub(n, one)
		n.And(n, temp)
	}
	return count
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
			internalKeys:       NewCache[InternalKeys](),
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

func (f *FetcherPegouts) getInternalKey(typeId int) ([]byte, error) {
	query := `
    SELECT payload::json
    FROM metrics_data 
    WHERE type_id = $1
    ORDER BY id DESC
    LIMIT 1
`
	rows, err := f.db.Query(query, typeId)
	if err != nil {
		return []byte{}, err
	}

	defer rows.Close()

	var data string
	if rows.Next() {
		err = rows.Scan(
			&data,
		)
		if err != nil {
			return []byte{}, err
		}
	}

	var dkgData coordinator.DKG
	err = json.Unmarshal([]byte(data), &dkgData)
	if err != nil {
		return []byte{}, err
	}

	return dkgData.R3.Data.InternalKey, nil
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

	now := time.Now()
	for _, pegout := range pegouts {
		if pegout.IsAutopegout {
			f.pegoutTempData.lastAutopegoutDate = now
		}
		if now.After(f.pegoutTempData.lastAutopegoutDate.Add(AUTOPEGOUT_MAX_DELAY)) {
			autopegoutDelayed.Set(1)
		} else {
			autopegoutDelayed.Set(0)
		}
	}
}

func (f *FetcherPegouts) setWrongInternalKeyMetric(
	pegouts map[uint64]coordinator.PegoutRecord,
	internalKeys InternalKeys,
) {
	for _, pegout := range pegouts {
		if !bytes.Equal(pegout.InternalKey, internalKeys.dkg) ||
			!bytes.Equal(pegout.InternalKey, internalKeys.prevDkg) {
			wrongInternalKey.WithLabelValues(string(pegout.InternalKey), utils.AddrToRawString(pegout.PegoutAddress)).Set(1)
		} else {
			wrongInternalKey.WithLabelValues(string(pegout.InternalKey), utils.AddrToRawString(pegout.PegoutAddress)).Set(0)
		}
	}
}

func (f *FetcherPegouts) setInsufficientValidatorsMetric(pegouts map[uint64]coordinator.PegoutRecord) {
	for _, pegout := range pegouts {
		if pegout.MaxSigners < EXPECTED_SIGNERS_COUNT {
			insufficientValidators.Set(1)
		} else {
			insufficientValidators.Set(0)
		}
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

	fmt.Println(f.pegoutTempData.unsignedPegouts)

	if unsignedPegouts == nil {
		logger.Log.Debug().Msg("FetcherPegouts: Contract returns unsignedPegouts is null")
	}

	unsignedPegoutsLen.WithLabelValues("Unsigned pegouts length").Set(float64(len(unsignedPegouts)))

	f.setDelayedMetric(unsignedPegouts)
	f.setAutopegoutDelayedMetric(unsignedPegouts)
	f.setInsufficientValidatorsMetric(unsignedPegouts)
	f.setSigningRestartMetric(unsignedPegouts)
	f.setRestartSigningPegoutCountMetric(f.pegoutTempData.expirations)
	f.setSigningMaskMetrics(unsignedPegouts[0])
	f.setCulpritMetrics(unsignedPegouts[0])
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

	var internalKeys InternalKeys
	internalKeysCache, ok := f.pegoutTempData.internalKeys.Get("InternalKeys")
	if ok {
		internalKeys = internalKeysCache
	} else {
		dkg, err := f.getInternalKey(0)
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherPegouts").
				Msg("fetch failed")
		}

		prevDkg, err := f.getInternalKey(1)
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherPegouts").
				Msg("fetch failed")
		}
		internalKeys = InternalKeys{dkg, prevDkg}
		f.pegoutTempData.internalKeys.Set("InternalKeys", internalKeys, time.Duration(f.period)*time.Second)
	}

	f.setWrongInternalKeyMetric(unsignedPegouts, internalKeys)
	f.pegoutTempData.unsignedPegouts.DeleteExpired()
	f.pegoutTempData.signedPegouts.DeleteExpired()
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
