package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/data_models"
)

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

	contractTeleportData := data_models.ContractTeleportStorageDbRow{
		Id:                   storage.Id,
		TeleportAddress:      utils.AddrToRawString(fetcher.teleportContract.Addr),
		MinterAddress:        utils.AddrToRawString(storage.MinterAddress),
		BitcoinClientAddress: utils.AddrToRawString(storage.BitcoinClientAddress),
		CoordinatorAddress:   utils.AddrToRawString(storage.CoordinatorAddress),
		InspectorAddress:     utils.AddrToRawString(storage.InspectorAddress),
		ConfiguratorAddress:  utils.AddrToRawString(storage.ConfiguratorAddress),
		TweakedPubkey:        storage.TweakedPubkey,
		InternalKey:          storage.InternalKey,
		NextSVB:              storage.NextSVB,
		BaseSVB:              storage.BaseSVB,
		PegoutChainCounter:   storage.PegoutChainCounter,
		LastPegoutTxID:       storage.LastPegoutTxID,
		CsvLock:              storage.CsvLock,
		Limits:               storage.Limits,
		TotalServiceFee:      storage.TotalServiceFee,
		Enabled:              storage.Enabled,
		PeginsCount:          ConvertDeposits(storage.Deposits),
		UTXOset:              ConvertUTXOSet(storage.UTXOset),
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

func ConvertDeposits(data map[uint64]teleportcontract.DepositData) int32 {
	return int32(len(data))
}

func ConvertUTXOSet(utxoSet map[string]teleportcontract.UTXOData) *[]data_models.ContractTeleportUTXO {
	contractTeleportUTXOData := []data_models.ContractTeleportUTXO{}

	for address, utxo := range utxoSet {
		cutxo := data_models.ContractTeleportUTXO{
			Address:       address,
			Amount:        utxo.Amount,
			Index:         utxo.Index,
			TapMerkleRoot: utxo.TapMerkleRoot,
			MintAddress:   utils.AddrToRawString(utxo.MintAddress),
			Script:        utxo.Script,
		}

		contractTeleportUTXOData = append(contractTeleportUTXOData, cutxo)
	}

	return &contractTeleportUTXOData
}
