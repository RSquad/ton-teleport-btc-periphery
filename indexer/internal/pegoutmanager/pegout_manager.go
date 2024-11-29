package pegoutmanager

import (
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
	teleportContract  *teleportcontract.TeleportContract
	tonClient         *tonclient.TonClient
	tonCenterV3Client *toncenterv3.Client
}

func New(
	ctx context.Context,
	repo *ent.Client,
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
		teleportContract:  teleportContract,
		tonClient:         tonClient,
		tonCenterV3Client: tonCenterV3Client,
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
		Where(entpegout.StatusNEQ(entpegout.StatusCompleted)).
		Limit(128).
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
		SetStatus(entpegout.StatusCompleted).
		SetBitcoinTxRaw(txHex).
		SetBitcoinTxId(pegoutTx.TxID()).
		Where(entpegout.ID(pegout.ID)).
		Exec(c.ctx)
	if err != nil {
		return fmt.Errorf("failed to update pegout: %w", err)
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
	for key := range *txParts.Inputs {
		inputTxIDs = append(inputTxIDs, key)
	}
	sort.Slice(inputTxIDs, func(i, j int) bool {
		return inputTxIDs[i] > inputTxIDs[j]
	})

	for _, inputTxID := range inputTxIDs {
		input, ok := (*txParts.Inputs)[inputTxID]
		if !ok {
			return nil, fmt.Errorf("missing input: %v", inputTxID)
		}
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
		packet.Inputs = append(packet.Inputs, pInput)
	}

	for i := range inputTxIDs {
		signature, ok := (*txParts.Signatures)[strconv.Itoa(i)]
		if !ok {
			return nil, fmt.Errorf("missing signature for input index %d", i)
		}
		packet.Inputs[i].TaprootKeySpendSig = signature
	}

	err = psbt.MaybeFinalizeAll(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to finalize inputs: %w", err)
	}

	finalizedTx, err := psbt.Extract(packet)
	if err != nil {
		return nil, fmt.Errorf("failed to extract finalized tx: %w", err)
	}

	return finalizedTx, nil
}
