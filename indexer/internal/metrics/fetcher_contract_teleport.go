package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/xssnick/tonutils-go/address"
)

type ContractTeleportData struct {
	Id                   uint16
	MinterAddress        *address.Address
	BitcoinClientAddress *address.Address
	CoordinatorAddress   *address.Address
	InspectorAddress     *address.Address
	ConfiguratorAddress  *address.Address
	TweakedPubkey        string
	InternalKey          string
	Deposits             map[uint64]teleportcontract.DepositData
	NextSVB              uint16
	BaseSVB              uint16
	PegoutChainCounter   uint64
	LastPegoutTxID       *chainhash.Hash
	CsvLock              uint32
	Limits               teleportcontract.Limits
	TotalServiceFee      int32
	Enabled              bool
	UTXOset              map[string]teleportcontract.UTXOData
}

type FetcherContractTeleport struct {
	chDB             chan PayloadDB
	teleportContract *teleportcontract.TeleportContract
	period           int64 // Fetch period (in seconds)
}

func NewFetcherContractTeleport(
	chDB chan PayloadDB,
	teleportContract *teleportcontract.TeleportContract,
	period int64,
) *FetcherContractTeleport {
	return &FetcherContractTeleport{
		chDB:             chDB,
		teleportContract: teleportContract,
		period:           period,
	}
}

func (fetcher *FetcherContractTeleport) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherContractTeleport: stopped")
	logger.DefaultLogStartWork("FetcherContractTeleport: starting...")

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("DKG Fetcher received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
		}
	}
}

func (fetcher *FetcherContractTeleport) Fetch() {
	storage, err := fetcher.teleportContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractTeleport: failed to retrieve storage cell, error: %v", err))
		return
	}

	contractTeleportData := ContractTeleportData{
		Id:                   storage.Id,
		MinterAddress:        storage.MinterAddress,
		BitcoinClientAddress: storage.BitcoinClientAddress,
		CoordinatorAddress:   storage.CoordinatorAddress,
		InspectorAddress:     storage.InspectorAddress,
		ConfiguratorAddress:  storage.ConfiguratorAddress,
		TweakedPubkey:        storage.TweakedPubkey,
		InternalKey:          storage.InternalKey,
		Deposits:             storage.Deposits,
		NextSVB:              storage.NextSVB,
		BaseSVB:              storage.BaseSVB,
		PegoutChainCounter:   storage.PegoutChainCounter,
		LastPegoutTxID:       storage.LastPegoutTxID,
		CsvLock:              storage.CsvLock,
		Limits:               storage.Limits,
		TotalServiceFee:      storage.TotalServiceFee,
		Enabled:              storage.Enabled,
		UTXOset:              storage.UTXOset,
	}

	jsonData, err := json.Marshal(contractTeleportData)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractTeleport").
			Msg("failed to serialize FetcherContractTeleport->json")
	}

	fetcher.chDB <- PayloadDB{
		timestamp: time.Now(),
		typeId:    PayloadTypeContractTeleport,
		payload:   string(jsonData),
	}
}
