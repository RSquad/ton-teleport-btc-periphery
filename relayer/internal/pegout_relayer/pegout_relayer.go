package pegoutrelayer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
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
	return &PegoutRelayer{
		bitcoinClient:         bitcoinClient,
		teleportContract:      teleportContract,
		bitcoinClientContract: bitcoinClientContract,
		isRelaying:            false,
	}, nil
}

func (c *PegoutRelayer) Relay() error {
	if c.isRelaying {
		return nil
	}

	c.isRelaying = true
	defer func() { c.isRelaying = false }()

	log.Println("[PegoutRelayer] relay started")

	teleportContractStorage, err := c.teleportContract.GetStorage(nil)
	if err != nil {
		return err
	}

	if teleportContractStorage.PegoutChainCounter == 0 {
		log.Println("[PegoutRelayer] nothing to relay")
		return nil
	}

	blockHash, err := c.bitcoinClient.GetBlockHashByTxID(teleportContractStorage.LastPegoutTxID)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get last pegout tx block hash: %w", err)
	}

	blockHeight, err := c.bitcoinClient.GetBlockHeightByHash(blockHash)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get last pegout tx block height: %w", err)
	}

	lastConfirmedBlockHash, err := c.bitcoinClientContract.GetLastConfirmedBlockHash()
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get bitcoin client contract last confirmed block hash: %w", err)
	}

	lastConfirmedBlockHeight, err := c.bitcoinClient.GetBlockHeightByHash(lastConfirmedBlockHash)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get bitcoin client contract last confirmed block height: %w", err)
	}

	confirmationsNeeded, err := c.bitcoinClientContract.GetConfirmationsNeeded()
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get bitcoin client contract confirmations needed: %w", err)
	}

	if lastConfirmedBlockHeight-blockHeight >= confirmationsNeeded {
		txProof, err := c.getTxProof(teleportContractStorage.LastPegoutTxID, blockHash)
		if err != nil {
			return fmt.Errorf("[PegoutRelayer] failed to get last pegout tx proof: %w", err)
		}

		merkleBlock, err := c.decodeTxProof(txProof)
		if err != nil {
			return fmt.Errorf("[PegoutRelayer] failed to decode last pegout tx proof: %w", err)
		}

		log.Printf("[PegoutRelayer] sending last pegout tx proof: txId=%v", teleportContractStorage.LastPegoutTxID.String())

		tx, _, err := c.teleportContract.SendPegoutProof(teleportContractStorage.LastPegoutTxID, blockHash, merkleBlock)
		if err != nil {
			return fmt.Errorf("[PegoutRelayer] failed to send pegout tx proof: %w", err)
		}

		log.Printf("[PegoutRelayer] last pegout tx proof sent: tonTxHash=%s", hex.EncodeToString(tx.Hash))

		return nil
	}

	log.Println("[PegoutRelayer] last pegout tx not confirmed yet")

	return nil
}

func (c *PegoutRelayer) decodeTxProof(txProof []byte) (*wire.MsgMerkleBlock, error) {
	var merkleBlock wire.MsgMerkleBlock
	buf := bytes.NewBuffer(txProof)
	err := merkleBlock.BtcDecode(buf, wire.ProtocolVersion, wire.BaseEncoding)
	if err != nil {
		return nil, err
	}

	return &merkleBlock, nil
}

func (c *PegoutRelayer) getTxProof(lastPegoutTxID *chainhash.Hash, blockHash *chainhash.Hash) (
	[]byte,
	error,
) {
	txProofStr, err := c.bitcoinClient.GetTxProof(
		lastPegoutTxID,
		blockHash,
	)
	if err != nil {
		return nil, err
	}

	txProof, err := hex.DecodeString(txProofStr)
	if err != nil {
		return nil, err
	}

	return txProof, nil
}
