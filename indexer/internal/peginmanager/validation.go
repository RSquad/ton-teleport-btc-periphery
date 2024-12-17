package peginmanager

import (
	"fmt"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/internalkey"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegin"
)

func (c *PeginManager) InternalKeyExists(
	internalKey string,
) (bool, error) {
	exists, err := c.repo.InternalKey.Query().Where(internalkey.KeyEQ(internalKey)).Exist(c.ctx)
	return exists, err
}

func (c *PeginManager) BitcoinTxExists(
	bitcoinTxIdStr string,
) (bool, *btcjson.TxRawResult, error) {
	bitcoinTxId, err := chainhash.NewHashFromStr(bitcoinTxIdStr)
	if err != nil {
		return false, nil, fmt.Errorf("invalid bitcoin tx id: %w", err)
	}
	bitcoinTx, err := c.bitcoinClient.RPCClient.GetRawTransactionVerbose(bitcoinTxId)
	if err != nil {
		return false, nil, fmt.Errorf("bitcoin tx not found: %w", err)
	}
	return true, bitcoinTx, nil
}

func (c *PeginManager) PeginExists(
	bitcoinTxIdStr string,
) (bool, error) {
	exists, err := c.repo.Pegin.Query().Where(pegin.BitcoinTxIdEQ(bitcoinTxIdStr)).Exist(c.ctx)
	return exists, err
}
