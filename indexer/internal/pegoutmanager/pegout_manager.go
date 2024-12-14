package pegoutmanager

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	entpegout "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegout"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	tonclient "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/ton_client"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/toncenterv3"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
)

type PegoutManager struct {
	ctx               context.Context
	repo              *ent.Client
	bitcoinClient     *bitcoin.Client
	tonClient         *tonclient.TonClient
	tonCenterV3Client *toncenterv3.Client
	teleportContract  *teleportcontract.TeleportContract
}

func New(
	ctx context.Context,
	repo *ent.Client,
	bitcoinClient *bitcoin.Client,
	tonClient *tonclient.TonClient,
	tonCenterV3Client *toncenterv3.Client,
	teleportContract *teleportcontract.TeleportContract,
) (
	*PegoutManager,
	error,
) {
	pegoutManager := &PegoutManager{
		ctx:               ctx,
		repo:              repo,
		bitcoinClient:     bitcoinClient,
		tonClient:         tonClient,
		tonCenterV3Client: tonCenterV3Client,
		teleportContract:  teleportContract,
	}

	return pegoutManager, nil
}

func (pm *PegoutManager) Run() {
	const sleepDuration = 3 * time.Second
	for {
		if err := pm.processPegouts(); err != nil {
			log.Printf("failed to process pegouts: %v", err)
		}
		time.Sleep(sleepDuration)
	}
}

func (c *PegoutManager) processPegouts() error {
	pegouts, err := c.repo.Pegout.Query().
		Where(entpegout.StatusNEQ(entpegout.StatusConfirmed)).
		Limit(512).
		All(c.ctx)
	if err != nil || len(pegouts) == 0 {
		return err
	}

	block, err := c.tonClient.API.CurrentMasterchainInfo(c.ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, pegout := range pegouts {
		wg.Add(1)
		go func(pegout *ent.Pegout) {
			defer wg.Done()
			if err := c.processPegout(block, pegout); err != nil {
				log.Printf("failed to process pegout id=%v addr=%v: %v", pegout.ID, pegout.Addr, err)
			}
		}(pegout)
	}

	wg.Wait()
	return nil
}

func (c *PegoutManager) processPegout(
	block *ton.BlockIDExt,
	pegout *ent.Pegout,
) error {
	switch pegout.Status {
	case entpegout.StatusSigning:
		return c.handleSigningPegout(block, pegout)
	case entpegout.StatusSigned:
		return c.handleSignedPegout(pegout)
	}
	return nil
}

func (c *PegoutManager) handleSigningPegout(
	block *ton.BlockIDExt,
	pegout *ent.Pegout,
) error {
	pegoutState, err := c.tonClient.API.GetAccount(c.ctx, block, address.MustParseRawAddr(pegout.Addr))
	if err != nil {
		return fmt.Errorf("failed to get pegout state: %w", err)
	}
	if !pegoutState.IsActive {
		return fmt.Errorf("pegout contract is not active")
	}

	pegoutContract := pegoutcontract.New(
		address.MustParseRawAddr(pegout.Addr),
		c.tonClient,
		c.ctx,
	)

	pegoutTxParts, err := pegoutContract.GetTxParts(block)
	if err != nil {
		return fmt.Errorf("failed to get pegout tx parts: %w", err)
	}

	pegoutTx, err := c.buildPegoutTx(pegoutTxParts)
	if err != nil {
		return fmt.Errorf("failed to build pegout tx: %w", err)
	}

	txHex, err := bitcoin.TxToHex(pegoutTx)
	if err != nil {
		return fmt.Errorf("failed to serialize pegout tx: %w", err)
	}

	err = c.repo.Pegout.Update().
		SetStatus(entpegout.StatusSigned).
		SetBitcoinTxRaw(txHex).
		SetBitcoinTxId(pegoutTx.TxID()).
		Where(entpegout.ID(pegout.ID)).
		Exec(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to update pegout: %w", err)
	}

	return nil
}

func (c *PegoutManager) handleSignedPegout(
	pegout *ent.Pegout,
) error {
	txHash, err := chainhash.NewHashFromStr(pegout.BitcoinTxId)
	if err != nil {
		return fmt.Errorf("failed to parse tx hash: %w", err)
	}
	txVerbose, err := c.bitcoinClient.RPCClient.GetRawTransactionVerbose(txHash)
	if err == nil {
		err = c.repo.Pegout.Update().
			SetStatus(entpegout.StatusConfirmed).
			SetBitcoinBlockHash(txVerbose.BlockHash).
			Where(entpegout.ID(pegout.ID)).
			Exec(c.ctx)
		if err != nil {
			return fmt.Errorf("failed to update pegout: %w", err)
		}
		return nil
	}

	tx, err := bitcoin.HexToTx(pegout.BitcoinTxRaw, 2)
	if err != nil {
		return fmt.Errorf("failed to serialize pegout tx: %w", err)
	}
	_, err = c.bitcoinClient.RPCClient.SendRawTransaction(tx, false)
	if err != nil {
		return fmt.Errorf("failed to send pegout tx: %w", err)
	}

	return nil
}

func (c *PegoutManager) buildPegoutTx(txParts *pegoutcontract.TxParts) (*wire.MsgTx, error) {
	packet, err := psbt.NewFromUnsignedTx(wire.NewMsgTx(2))
	if err != nil {
		return nil, err
	}

	for _, output := range txParts.Outputs {
		txOut := wire.NewTxOut(int64(output.Amount), output.BitcoinScript)
		packet.UnsignedTx.AddTxOut(txOut)
		packet.Outputs = append(packet.Outputs, psbt.POutput{})
	}

	inputTxIDs := make([]string, 0, len(*txParts.Inputs))
	for txid := range *txParts.Inputs {
		inputTxIDs = append(inputTxIDs, txid)
	}
	sort.Slice(inputTxIDs, func(i, j int) bool {
		txidIBytes, errI := hex.DecodeString(inputTxIDs[i])
		txidJBytes, errJ := hex.DecodeString(inputTxIDs[j])
		if errI != nil || errJ != nil {
			return false
		}
		return bytes.Compare(txidIBytes, txidJBytes) < 0
	})

	for i, inputTxID := range inputTxIDs {
		input := (*txParts.Inputs)[inputTxID]
		keyBytes, err := hex.DecodeString(inputTxID)
		keyBytes = utils.BytesPadTo(keyBytes, 32)
		if err != nil {
			return nil, fmt.Errorf("error decoding key: %v", err)
		}
		hash, err := chainhash.NewHash(keyBytes)
		if err != nil {
			return nil, fmt.Errorf("invalid previous tx hash: %v", err)
		}

		txIn := wire.NewTxIn(&wire.OutPoint{
			Hash:  *hash,
			Index: uint32(input.Index),
		}, nil, nil)
		packet.UnsignedTx.AddTxIn(txIn)

		pInput := psbt.PInput{
			WitnessUtxo: &wire.TxOut{
				Value:    input.Amount.Int64(),
				PkScript: input.BitcoinScript,
			},
			TaprootInternalKey: txParts.InternalKey,
		}
		if len(input.BitcoinMerkleRoot) > 0 {
			pInput.TaprootMerkleRoot = input.BitcoinMerkleRoot
		}
		signature := (*txParts.Signatures)[strconv.Itoa(i)]
		if len(signature) > 64 {
			signature = signature[len(signature)-64:]
		}
		pInput.TaprootKeySpendSig = signature
		packet.Inputs = append(packet.Inputs, pInput)
	}

	if err := psbt.MaybeFinalizeAll(packet); err != nil {
		return nil, fmt.Errorf("failed to finalize inputs: %w", err)
	}

	finalizedTx, err := psbt.Extract(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to extract finalized tx: %w", err)
	}

	return finalizedTx, nil
}
