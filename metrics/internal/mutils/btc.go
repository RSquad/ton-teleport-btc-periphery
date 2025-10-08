package mutils

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
)

func GetBestBlockHeight(bitcoinClient *bitcoin.Client) (int, error) {
	btcInfo, err := bitcoinClient.GetBlockChainInfo()
	if err != nil {
		return 0, fmt.Errorf("failed to get btc blockchain info: %v", err)
	}

	return int(btcInfo.Blocks), nil
}

func GetCPFPChainSize(bitcoinClient *bitcoin.Client, txHash *chainhash.Hash) (int, error) {
	chainSize := 0
	currentTxHash := txHash
	visited := make(map[string]bool) // Prevent infinite loops

	for {
		if visited[currentTxHash.String()] {
			break
		}
		visited[currentTxHash.String()] = true

		txResult, err := bitcoinClient.GetRawTransactionVerbose(currentTxHash)
		if err != nil {
			return 0, fmt.Errorf("failed to get transaction: %v", err)
		}

		if txResult.Confirmations > 0 {
			// Transaction is confirmed, not in mempool
			break
		}

		chainSize++

		// Find unconfirmed parent transactions
		hasUnconfirmedParent := false
		var parentTxHash *chainhash.Hash

		for _, vin := range txResult.Vin {
			if vin.Txid == "" {
				// Coinbase transaction
				continue
			}

			// Get parent transaction
			parentHash, err := chainhash.NewHashFromStr(vin.Txid)
			if err != nil {
				continue
			}

			parentTx, err := bitcoinClient.GetRawTransactionVerbose(parentHash)
			if err != nil {
				continue
			}

			// Check if parent is also unconfirmed (in mempool)
			if parentTx.Confirmations == 0 {
				hasUnconfirmedParent = true
				parentTxHash, err = chainhash.NewHashFromStr(vin.Txid)
				if err != nil {
					return 0, err
				}
				break // We found an unconfirmed parent, follow this chain
			}
		}

		if !hasUnconfirmedParent {
			// No more unconfirmed parents, we've reached the top of the chain
			break
		}

		// Move to the parent transaction
		currentTxHash = parentTxHash
	}

	return chainSize, nil
}
