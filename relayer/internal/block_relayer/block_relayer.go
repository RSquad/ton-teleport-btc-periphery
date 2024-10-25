package blockrelayer

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type BlockRelayer struct {
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *ton.BitcoinClientContract
	isRelaying            bool
	confirmationsNeeded   int64
}

func NewBlockRelayer(bitcoinClient *bitcoin.Client, bitcoinClientContract *ton.BitcoinClientContract) (
	*BlockRelayer,
	error,
) {
	confirmationsNeeded, err := bitcoinClientContract.GetConfirmationsNeeded()
	if err != nil {
		return nil, err
	}

	return &BlockRelayer{
		bitcoinClient:         bitcoinClient,
		bitcoinClientContract: bitcoinClientContract,
		confirmationsNeeded:   confirmationsNeeded,
	}, nil
}

func (c *BlockRelayer) Relay() error {
	if c.isRelaying {
		return nil
	}

	c.isRelaying = true
	defer func() { c.isRelaying = false }()

	log.Println("[BlockRelayer] relay started")

	bitcoinHeight, err := c.bitcoinClient.RPCClient.GetBlockCount()
	if err != nil {
		return fmt.Errorf("[BlockRelayer] failed to get bitcoin height: %w", err)
	}

	lastConfirmedBlockHash, err := c.bitcoinClientContract.GetLastConfirmedBlockHash()
	if err != nil {
		return fmt.Errorf("[BlockRelayer] failed to get last confirmed block hash: %w", err)
	}

	lastConfirmedBlockHeight, err := c.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		return fmt.Errorf("[BlockRelayer] failed to get last confirmed block height: %w", err)
	}

	candidateBlocksCount := c.calcCandidateBlocksCount(bitcoinHeight, lastConfirmedBlockHeight)

	log.Printf(
		"[BlockRelayer] bitcoin client contract sync state: outOfSync=%d bitcoinHeight=%d lastConfirmedBlockHeight=%d",
		bitcoinHeight-lastConfirmedBlockHeight-c.confirmationsNeeded+1,
		bitcoinHeight,
		lastConfirmedBlockHeight,
	)

	if candidateBlocksCount > 0 {
		registeredCandidateBlockHashes, err := c.bitcoinClientContract.GetCandidateBlockHashes()
		if err != nil {
			return fmt.Errorf("[BlockRelayer] failed to get registered candidate block hashes: %w", err)
		}

		candidateBlockHashes, err := c.getCandidateBlockHashes(lastConfirmedBlockHeight, candidateBlocksCount)
		if err != nil {
			return fmt.Errorf("[BlockRelayer] failed to get candidate block hashes: %w", err)
		}

		for _, candidateBlockHash := range candidateBlockHashes {
			if !bitcoin.SliceOfHashesContains(registeredCandidateBlockHashes, candidateBlockHash) {
				candidateBlockHeaderToSend, err := c.bitcoinClient.RPCClient.GetBlockHeader(candidateBlockHash)
				if err != nil {
					return fmt.Errorf("[BlockRelayer] failed to get candidate block header: %w", err)
				}

				log.Printf(
					"[BlockRelayer] sending new candidate block header: blockHash=%v",
					candidateBlockHeaderToSend.BlockHash().String(),
				)

				tx, _, err := c.bitcoinClientContract.SendCandidateBlockHeader(candidateBlockHeaderToSend)
				if err != nil {
					return fmt.Errorf("[BlockRelayer] failed to send candidate block header: %w", err)
				}

				log.Printf("[BlockRelayer] candidate block header sent: txHash=%v", utils.BytesToHexString(tx.Hash))

				break
			}
		}
	}

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
