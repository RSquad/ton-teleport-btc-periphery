package pegoutrelayer

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	bitcoinclientcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
)

type PegoutRelayer struct {
	bitcoinClient         *bitcoin.Client
	teleportContract      *teleportcontract.TeleportContract
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	isRelaying            bool
}

func NewPegoutRelayer(
	bitcoinClient *bitcoin.Client,
	teleportContract *teleportcontract.TeleportContract,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
) (
	*PegoutRelayer,
	error,
) {
	component := "PegoutRelayer"

	logger.Log.Info().
		Str("component", component).
		Msg("Pegout relayer created")

	return &PegoutRelayer{
		bitcoinClient:         bitcoinClient,
		teleportContract:      teleportContract,
		bitcoinClientContract: bitcoinClientContract,
		isRelaying:            false,
	}, nil
}

func (c *PegoutRelayer) Relay() error {
	component := "PegoutRelayer"

	if c.isRelaying {
		logger.Log.Debug().
			Str("component", component).
			Msg("Already relaying, skipping")
		return nil
	}

	c.isRelaying = true
	defer func() { c.isRelaying = false }()

	logger.Log.Info().
		Str("component", component).
		Msg("Relay started")

	// Get teleport contract storage
	teleportContractStorage, err := c.teleportContract.GetStorage(nil)
	if err != nil {
		return fmt.Errorf("failed to get teleport contract storage: %w", err)
	}

	// Check if there are pegouts to relay
	if teleportContractStorage.PegoutChainCounter == 0 {
		logger.Log.Info().
			Str("component", component).
			Int64("pegout_chain_counter", 0).
			Msg("No pegouts to relay")
		return nil
	}

	logger.Log.Debug().
		Str("component", component).
		Str("last_pegout_tx_id", teleportContractStorage.LastPegoutTxID.String()).
		Msg("Found pegouts to relay")

	// Get block hash for the last pegout transaction
	blockHash, err := c.bitcoinClient.GetBlockHashByTxID(teleportContractStorage.LastPegoutTxID)
	if err != nil {
		return fmt.Errorf("failed to get last pegout tx block hash: %w. LastPegoutTxID: '%s'", err, teleportContractStorage.LastPegoutTxID.String())
	}

	// Get block height for the block hash
	blockHeight, err := c.bitcoinClient.GetBlockHeightByHash(blockHash)
	if err != nil {
		return fmt.Errorf("failed to get last pegout tx block height: %w", err)
	}

	// Get last confirmed block from Bitcoin client contract
	lastConfirmedBlockHash, err := c.bitcoinClientContract.GetLastConfirmedBlockHash()
	if err != nil {
		return fmt.Errorf("failed to get bitcoin client contract last confirmed block hash: %w", err)
	}

	// Get height for last confirmed block
	lastConfirmedBlockHeight, err := c.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		return fmt.Errorf("failed to get bitcoin client contract last confirmed block height: %w", err)
	}

	logger.Log.Info().
		Str("component", component).
		Int64("pegout_block_height", blockHeight).
		Int64("last_confirmed_block_height", lastConfirmedBlockHeight).
		Str("last_pegout_tx_id", teleportContractStorage.LastPegoutTxID.String()).
		Msg("Checking pegout confirmation status")

	if blockHeight <= lastConfirmedBlockHeight {
		// Get transaction proof
		txProof, err := c.getTxProof(teleportContractStorage.LastPegoutTxID, blockHash)
		if err != nil {
			return fmt.Errorf("failed to get last pegout tx proof: %w", err)
		}

		// Decode merkle block proof
		merkleBlock, err := c.decodeTxProof(txProof)
		if err != nil {
			return fmt.Errorf("failed to decode last pegout tx proof: %w", err)
		}

		logger.Log.Info().
			Str("component", component).
			Str("tx_id", teleportContractStorage.LastPegoutTxID.String()).
			Str("block_hash", blockHash.String()).
			Int("proof_size_bytes", len(txProof)).
			Msg("Sending pegout transaction proof")

		// Send proof to teleport contract
		tx, _, err := c.teleportContract.SendPegoutProof(teleportContractStorage.LastPegoutTxID, blockHash, merkleBlock)
		if err != nil {
			return fmt.Errorf("failed to send pegout tx proof: %w", err)
		}

		logger.Log.Info().
			Str("component", component).
			Str("tx_id", teleportContractStorage.LastPegoutTxID.String()).
			Str("ton_tx_hash", hex.EncodeToString(tx.Hash)).
			Str("block_hash", blockHash.String()).
			Msg("Pegout transaction proof sent successfully")

		return nil
	}

	logger.Log.Info().
		Str("component", component).
		Str("tx_id", teleportContractStorage.LastPegoutTxID.String()).
		Int64("pegout_block_height", blockHeight).
		Int64("last_confirmed_height", lastConfirmedBlockHeight).
		Int64("confirmations_needed", lastConfirmedBlockHeight-blockHeight).
		Msg("Pegout transaction not confirmed yet")

	return nil
}

func (c *PegoutRelayer) decodeTxProof(txProof []byte) (*wire.MsgMerkleBlock, error) {
	component := "PegoutRelayer"

	var merkleBlock wire.MsgMerkleBlock
	buf := bytes.NewBuffer(txProof)
	err := merkleBlock.BtcDecode(buf, wire.ProtocolVersion, wire.BaseEncoding)
	if err != nil {
		logger.Log.Error().
			Str("component", component).
			Int("proof_size_bytes", len(txProof)).
			Err(err).
			Msg("Failed to decode merkle block proof")
		return nil, err
	}

	logger.Log.Debug().
		Str("component", component).
		Int("proof_size_bytes", len(txProof)).
		Msg("Merkle block proof decoded successfully")

	return &merkleBlock, nil
}

func (c *PegoutRelayer) getTxProof(lastPegoutTxID *chainhash.Hash, blockHash *chainhash.Hash) (
	[]byte,
	error,
) {
	component := "PegoutRelayer"

	txProofStr, err := c.bitcoinClient.GetTxProof(
		lastPegoutTxID,
		blockHash,
	)
	if err != nil {
		logger.Log.Error().
			Str("component", component).
			Str("tx_id", lastPegoutTxID.String()).
			Str("block_hash", blockHash.String()).
			Err(err).
			Msg("Failed to get transaction proof")
		return nil, err
	}

	txProof, err := hex.DecodeString(txProofStr)
	if err != nil {
		logger.Log.Error().
			Str("component", component).
			Str("tx_id", lastPegoutTxID.String()).
			Str("block_hash", blockHash.String()).
			Int("proof_string_length", len(txProofStr)).
			Err(err).
			Msg("Failed to decode hex transaction proof")
		return nil, err
	}

	logger.Log.Debug().
		Str("component", component).
		Str("tx_id", lastPegoutTxID.String()).
		Str("block_hash", blockHash.String()).
		Int("proof_size_bytes", len(txProof)).
		Msg("Transaction proof retrieved")

	return txProof, nil
}
