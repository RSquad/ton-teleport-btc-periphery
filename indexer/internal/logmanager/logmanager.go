package logmanager

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/ent"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/ent/burn"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/ent/mint"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/ent/reinit"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	teleportcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleport_contract"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type LogManager struct {
	ctx                         context.Context
	repo                        *ent.Client
	teleportContractLogParser   *teleportcontract.LogParser
	teleportContractLogListener *teleportcontract.LogListener
}

func New(
	ctx context.Context,
	repo *ent.Client,
	tonCenterV3Client *ton.TonCenterV3Client,
	teleportContractAddr *address.Address,
) (
	*LogManager,
	error,
) {
	logManager := &LogManager{
		ctx:  ctx,
		repo: repo,
	}

	teleportContractLogParser, err := teleportcontract.NewLogParser()
	if err != nil {
		return nil, fmt.Errorf("[LogManager] failed to create teleport contract log parser: %w", err)
	}
	logManager.teleportContractLogParser = teleportContractLogParser

	teleportContractLogListener, err := teleportcontract.NewLogListener(
		tonCenterV3Client,
		teleportContractAddr,
		logManager.onLogReceived,
	)
	if err != nil {
		return nil, fmt.Errorf("[LogManager] failed to create teleport contract log listener: %w", err)
	}
	logManager.teleportContractLogListener = teleportContractLogListener

	return logManager, nil
}

func (c *LogManager) Run() {
	c.teleportContractLogListener.StartListen()
}

func (c *LogManager) onLogReceived(logCell *cell.Cell, tonMsgHash string, createdAt time.Time) {
	parsedLog, err := c.teleportContractLogParser.Parse(logCell)
	if err != nil {
		log.Printf("[LogManager] failed to parse log %v", err)
	}
	// TODO: Need to add TonMsg model to check existence before saving logs to prevent duplicates
	switch typedParsedLog := parsedLog.(type) {
	case *teleportcontract.MintLog:
		exists, err := c.repo.Mint.Query().
			Where(mint.TonMsgHash(tonMsgHash)).
			Exist(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to check mint existence: %v", err)
			return
		}
		if exists {
			log.Printf("[LogManager] mint with tonMsgHash %s already exists, skipping", tonMsgHash)
			return
		}

		_, err = c.repo.Mint.Create().
			SetReceiverAddr(typedParsedLog.ReceiverAddr.String()).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetTonMsgHash(tonMsgHash).
			SetCreatedAt(createdAt).
			Save(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to save mint: %v", err)
		}

	case *teleportcontract.BurnLog:
		exists, err := c.repo.Burn.Query().
			Where(burn.TonMsgHash(tonMsgHash)).
			Exist(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to check burn existence: %v", err)
			return
		}
		if exists {
			log.Printf("[LogManager] burn with tonMsgHash %s already exists, skipping", tonMsgHash)
			return
		}

		_, err = c.repo.Burn.Create().
			SetExternalId(uint64(typedParsedLog.ID)).
			SetSenderAddr(typedParsedLog.SenderAddr.String()).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
			SetTonMsgHash(tonMsgHash).
			SetCreatedAt(createdAt).
			Save(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to save burn: %v", err)
		}

	case *teleportcontract.ReinitLog:
		exists, err := c.repo.Reinit.Query().
			Where(reinit.TonMsgHash(tonMsgHash)).
			Exist(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to check reinit existence: %v", err)
			return
		}
		if exists {
			log.Printf("[LogManager] reinit with tonMsgHash %s already exists, skipping", tonMsgHash)
			return
		}

		_, err = c.repo.Reinit.Create().
			SetExternalId(uint64(typedParsedLog.ID)).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
			SetTonMsgHash(tonMsgHash).
			SetCreatedAt(createdAt).
			Save(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to save reinit: %v", err)
		}

	default:
		log.Printf("[LogManager] unknown log type %T\n", typedParsedLog.GetLogID())
	}
}
