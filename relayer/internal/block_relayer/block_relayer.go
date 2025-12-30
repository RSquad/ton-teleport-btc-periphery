package blockrelayer

import (
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	bitcoinclientcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/bitcoinclientcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type BlockRelayer struct {
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract
	isRelaying            bool
	confirmationsNeeded   int64
}

func NewBlockRelayer(
	bitcoinClient *bitcoin.Client,
	bitcoinClientContract *bitcoinclientcontract.BitcoinClientContract,
) (
	*BlockRelayer,
	error,
) {
	component := "BlockRelayer"

	confirmationsNeeded, err := bitcoinClientContract.GetConfirmationsNeeded()
	if err != nil {
		return nil, err
	}

	logger.Log.Info().
		Str("component", component).
		Int64("confirmations_needed", confirmationsNeeded).
		Msg("Block relayer created")

	return &BlockRelayer{
		bitcoinClient:         bitcoinClient,
		bitcoinClientContract: bitcoinClientContract,
		confirmationsNeeded:   confirmationsNeeded,
	}, nil
}

func (c *BlockRelayer) Relay() error {
	component := "BlockRelayer"

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

	// Get current Bitcoin height
	bitcoinHeight, err := c.bitcoinClient.RPCClient.GetBlockCount()
	if err != nil {
		return fmt.Errorf("failed to get bitcoin height: %w", err)
	}

	// Get last confirmed block hash from contract
	lastConfirmedBlockHash, err := c.bitcoinClientContract.GetLastConfirmedBlockHash()
	if err != nil {
		return fmt.Errorf("failed to get last confirmed block hash: %w", err)
	}

	// Get block height for the last confirmed hash
	lastConfirmedBlockHeight, err := c.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		return fmt.Errorf("failed to get last confirmed block height: %w", err)
	}

	// Calculate candidate blocks count
	candidateBlocksCount := c.calcCandidateBlocksCount(bitcoinHeight, lastConfirmedBlockHeight)

	if candidateBlocksCount > 0 {
		// Get registered candidate block hashes from contract
		registeredCandidateBlockHashes, err := c.bitcoinClientContract.GetCandidateBlockHashes()
		if err != nil {
			return fmt.Errorf("failed to get registered candidate block hashes: %w", err)
		}

		// Get candidate block hashes from Bitcoin network
		candidateBlockHashes, err := c.getCandidateBlockHashes(lastConfirmedBlockHeight, candidateBlocksCount)
		if err != nil {
			return fmt.Errorf("failed to get candidate block hashes: %w", err)
		}

		// Send new candidate blocks
		for _, candidateBlockHash := range candidateBlockHashes {
			if !bitcoin.SliceOfHashesContains(registeredCandidateBlockHashes, candidateBlockHash) {
				component := "BlockRelayer"

				blockHashStr := candidateBlockHash.String()

				logger.Log.Info().
					Str("component", component).
					Str("block_hash", blockHashStr).
					Msg("Sending new candidate block header")

				// Get block header
				candidateBlockHeaderToSend, err := c.bitcoinClient.RPCClient.GetBlockHeader(candidateBlockHash)
				if err != nil {
					return fmt.Errorf("failed to get candidate block header: %w", err)
				}

				// Send block header to contract

				logger.Log.Info().
					Str("component", component).
					Str("block_hash", candidateBlockHeaderToSend.BlockHash().String()).
					Msg("Sending new candidate block header")

				tx, _, err := c.bitcoinClientContract.SendCandidateBlockHeader(candidateBlockHeaderToSend)
				if err != nil {
					return fmt.Errorf("failed to send candidate block header: %w", err)
				}

				logger.Log.Info().
					Str("component", component).
					Str("block_hash", blockHashStr).
					Str("transaction_hash", hex.EncodeToString(tx.Hash)).
					Msg("Candidate block header sent successfully")

				break
			}
		}
	}

	logger.Log.Info().
		Str("component", component).
		Msg("Relay completed")

	return nil
}

func (c *BlockRelayer) calcCandidateBlocksCount(bitcoinHeight int64, lastConfirmedBlockHeight int64) int64 {
	return utils.MinInt(bitcoinHeight-lastConfirmedBlockHeight, c.confirmationsNeeded+1)
}

func (c *BlockRelayer) getCandidateBlockHashes(
	lastConfirmedBlockHeight int64,
	candidateBlocksCount int64,
) ([]*chainhash.Hash, error) {
	return c.bitcoinClient.GetBlockHashesByStartHeight(
		lastConfirmedBlockHeight+1,
		utils.MinInt(c.confirmationsNeeded+1, candidateBlocksCount),
	)
}

func (c *BlockRelayer) sendCandidateBlock(candidateBlockHash *chainhash.Hash) error {
	return nil
}
