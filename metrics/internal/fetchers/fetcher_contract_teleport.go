package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type ContractTeleportUTXO struct {
	Address       string
	Amount        *big.Int
	Index         uint8
	TapMerkleRoot *chainhash.Hash
	MintAddress   string
	Script        string
}

type ContractTeleportData struct {
	Id                   uint16
	TeleportAddress      string
	MinterAddress        string
	BitcoinClientAddress string
	CoordinatorAddress   string
	InspectorAddress     string
	ConfiguratorAddress  string
	TweakedPubkey        string
	InternalKey          string
	NextSVB              uint16
	BaseSVB              uint16
	PegoutChainCounter   uint64
	LastPegoutTxID       *chainhash.Hash
	CsvLock              uint32
	Limits               teleportcontract.Limits
	TotalServiceFee      int32
	Enabled              bool
	PeginsCount          int32
	UTXOset              *[]ContractTeleportUTXO
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

/*
func (fetcher *FetcherContractTeleport) setAutopegoutFeeMetric(autopegoutFee *big.Int) {
	if autopegoutFee == nil {
		autopegoutFeeGauge.Set(-1)
		return
	}
	autopegoutFeeGauge.Set(float64(autopegoutFee.Int64()))
}

func (fetcher *FetcherContractTeleport) setUtxoDifferentKeysMetric(utxo *[]ContractTeleportUTXO) {
	var prevKey *chainhash.Hash = (*utxo)[0].TapMerkleRoot
	for _, utxo := range *utxo {
		if utxo.TapMerkleRoot != prevKey {
			utxoKeysDifference.WithLabelValues(prevKey.String(), utxo.TapMerkleRoot.String()).Set(1)
			continue
		}
		utxoKeysDifference.WithLabelValues(prevKey.String(), utxo.TapMerkleRoot.String()).Set(0)
	}
}
*/

func (fetcher *FetcherContractTeleport) Fetch() {
	storage, err := fetcher.teleportContract.GetStorage(nil)
	if err != nil {
		logger.Log.Error().Msg(fmt.Sprintf("FetcherContractTeleport: failed to retrieve storage cell, error: %v", err))
		return
	}

	contractTeleportData := ContractTeleportData{
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

	/*
		autopegoutFee, err := fetcher.teleportContract.GetAutoPegoutFee(nil)
		if err != nil {
			logger.Log.Error().Err(err).
				Str("component", "FetcherContractTeleport").
				Msg("failed to get autopegout fee")
		}
	*/
	//var utxoLimit float64 = 252 // TODO: get limit value from teleport
	//utxoLimitGauge.Set(float64(utxoLimit))
	//utxoCountGauge.Set(float64(len(*contractTeleportData.UTXOset)))
	//totalSetrviceFeeGauge.Set(float64(contractTeleportData.TotalServiceFee))
	//fetcher.setAutopegoutFeeMetric(autopegoutFee)

	//fetcher.setUtxoDifferentKeysMetric(contractTeleportData.UTXOset)

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

func ConvertUTXOSet(utxoSet map[string]teleportcontract.UTXOData) *[]ContractTeleportUTXO {
	contractTeleportUTXOData := []ContractTeleportUTXO{}

	for address, utxo := range utxoSet {
		cutxo := ContractTeleportUTXO{
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
