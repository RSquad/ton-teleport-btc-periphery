package pegoutrelayer

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleport_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type PegoutRelayer struct {
	bitcoinClient    *bitcoin.Client
	teleportContract *teleportcontract.TeleportContract
	isRelaying       bool
}

func NewPegoutRelayer(
	bitcoinClient *bitcoin.Client,
	teleportContract *teleportcontract.TeleportContract,
) (
	*PegoutRelayer,
	error,
) {
	return &PegoutRelayer{
		bitcoinClient:    bitcoinClient,
		teleportContract: teleportContract,
		isRelaying:       false,
	}, nil
}

func (c *PegoutRelayer) Relay() error {
	if c.isRelaying {
		return nil
	}

	c.isRelaying = true
	defer func() { c.isRelaying = false }()

	log.Println("[PegoutRelayer] relay started")

	teleportContractStorage, err := c.teleportContract.GetStorage()
	if err != nil {
		return err
	}

	if teleportContractStorage.PegoutChainCounter == 0 {
		log.Println("[PegoutRelayer] nothing to relay")
		return nil
	}

	blockHash, err := c.bitcoinClient.GetBlockHashByTxID(teleportContractStorage.LastPegoutTxID)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get pegout tx block hash: %w", err)
	}

	txProof, err := c.getTxProof(teleportContractStorage.LastPegoutTxID, blockHash)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to get pegout tx proof: %w", err)
	}

	merkleBlock, err := c.decodeTxProof(txProof)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to decode pegout tx proof: %w", err)
	}

	log.Printf("[PegoutRelayer] sending pegout tx: txId=%v", teleportContractStorage.LastPegoutTxID.String())

	tx, _, err := c.teleportContract.SendPegoutProof(teleportContractStorage.LastPegoutTxID, blockHash, merkleBlock)
	if err != nil {
		return fmt.Errorf("[PegoutRelayer] failed to send pegout tx proof: %w", err)
	}

	log.Printf("[PegoutRelayer] pegout tx proof sent: tonTxHash=%v", utils.BytesToHexString(tx.Hash))

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
