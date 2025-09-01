package data_models

import (
	"math/big"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

type ContractTeleportUTXO struct {
	Address       string
	Amount        *big.Int
	Index         uint8
	TapMerkleRoot *chainhash.Hash
	MintAddress   string
	Script        string
}

type ContractTeleportStorageDbRow struct {
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
