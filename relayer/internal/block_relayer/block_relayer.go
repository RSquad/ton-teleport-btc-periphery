package blockrelayer

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"

	"lib/pkg/bitcoin"
	"lib/pkg/ton"
	utils2 "lib/pkg/utils"
)

type BlockRelayer struct {
	bitcoinClient         *bitcoin.Client
	bitcoinClientContract *ton.BitcoinClientContract
	isRelaying            bool
}

func NewBlockRelayer(bitcoinClient *bitcoin.Client, bitcoinClientContract *ton.BitcoinClientContract) (
	*BlockRelayer,
	error,
) {
	return &BlockRelayer{bitcoinClient: bitcoinClient, bitcoinClientContract: bitcoinClientContract}, nil
}

func (c *BlockRelayer) Relay() error {
	if c.isRelaying {
		return nil
	}

	c.isRelaying = true
	defer func() { c.isRelaying = false }()

	log.Println("block relay started")

	bitcoinHeight, err := c.bitcoinClient.RPCClient.GetBlockCount()
	if err != nil {
		return fmt.Errorf("failed to get bitcoin height: %v", err)
	}

	lastConfirmedBlockHash, err := c.bitcoinClientContract.GetLastConfirmedBlockHash()
	if err != nil {
		return fmt.Errorf("failed to get last confirmed block hash: %v", err)
	}

	lastConfirmedBlockHeight, err := c.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		return fmt.Errorf("failed to get last confirmed block height: %v", err)
	}

	candidateBlocksCount := c.calcCandidateBlocksCount(bitcoinHeight, lastConfirmedBlockHeight)

	log.Printf(
		"bitcoin client contract sync state: outOfSync=%d bitcoinHeight=%d lastConfirmedBlockHeight=%d",
		bitcoinHeight-lastConfirmedBlockHeight-bitcoin.ConfirmationsNeeded,
		bitcoinHeight,
		lastConfirmedBlockHeight,
	)

	if candidateBlocksCount > 0 {
		registeredCandidateBlockHashes, err := c.bitcoinClientContract.GetCandidateBlockHashes()
		if err != nil {
			return fmt.Errorf("failed to get registered candidate block hashes: %v", err)
		}

		candidateBlockHashes, err := c.getCandidateBlockHashes(lastConfirmedBlockHeight, candidateBlocksCount)
		if err != nil {
			return fmt.Errorf("failed to get candidate block hashes: %v", err)
		}

		for _, candidateBlockHash := range candidateBlockHashes {
			if !bitcoin.SliceOfHashesContains(registeredCandidateBlockHashes, candidateBlockHash) {
				candidateBlockHeaderToSend, err := c.bitcoinClient.RPCClient.GetBlockHeader(candidateBlockHash)
				if err != nil {
					return fmt.Errorf("failed to get candidate block header: %v", err)
				}

				log.Printf(
					"new candidate block found: blockHash=%v",
					candidateBlockHeaderToSend.BlockHash().String(),
				)

				tx, _, err := c.bitcoinClientContract.SendCandidateBlockHeader(candidateBlockHeaderToSend)
				if err != nil {
					return fmt.Errorf("failed to send candidate block header: %v", err)
				}

				log.Printf("candidate block header sent: txHash=%v", utils2.BytesToHexString(tx.Hash))

				break
			}
		}
	}

	return nil
}

func (c *BlockRelayer) calcCandidateBlocksCount(bitcoinHeight int64, lastConfirmedBlockHeight int64) int64 {
	return utils2.MinInt(bitcoinHeight-lastConfirmedBlockHeight, bitcoin.ConfirmationsNeeded)
}

func (c *BlockRelayer) getCandidateBlockHashes(
	lastConfirmedBlockHeight int64,
	candidateBlocksCount int64,
) ([]*chainhash.Hash, error) {
	return c.bitcoinClient.GetBlockHashesByStartHeight(
		lastConfirmedBlockHeight+1,
		utils2.MinInt(bitcoin.ConfirmationsNeeded, candidateBlocksCount),
	)
}
