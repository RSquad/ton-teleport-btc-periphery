package bitcoin

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
)

func BitcoinTxExists(
	bitcoinClient *bitcoin.Client,
	bitcoinTxIDStr string,
) (bool, *btcjson.TxRawResult, error) {
	bitcoinTxID, err := chainhash.NewHashFromStr(bitcoinTxIDStr)
	if err != nil {
		return false, nil, fmt.Errorf("invalid bitcoin tx id: %w", err)
	}
	bitcoinTx, err := bitcoinClient.GetRawTransactionVerbose(bitcoinTxID)
	if err != nil {
		return false, nil, fmt.Errorf("bitcoin tx not found: %w", err)
	}
	return true, bitcoinTx, nil
}
