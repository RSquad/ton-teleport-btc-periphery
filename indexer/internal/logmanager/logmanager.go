package logmanager

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/tonmsg"
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

func (c *LogManager) onLogReceived(logCell *cell.Cell, msgHash string, msgCreatedAt time.Time) {
	parsedLog, err := c.teleportContractLogParser.Parse(logCell)
	if err != nil {
		log.Printf("[LogManager] failed to parse log %v", err)
	}

	exists, err := c.repo.TonMsg.Query().
		Where(tonmsg.Hash(msgHash)).
		Exist(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to check ton msg existence: %v", err)
		return
	}
	if exists {
		log.Printf("[LogManager] msg with hash %s already exists, skipping", msgHash)
		return
	}

	tonMsg, err := c.repo.TonMsg.Create().
		SetHash(msgHash).
		SetCreatedAt(msgCreatedAt).
		Save(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to save ton msg: %v", err)
		return
	}

	switch typedParsedLog := parsedLog.(type) {
	case *teleportcontract.MintLog:
		_, err = c.repo.Mint.Create().
			SetReceiverAddr(typedParsedLog.ReceiverAddr.String()).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetTonMsg(tonMsg).
			Save(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to save mint: %v", err)
		}

	case *teleportcontract.BurnLog:
		_, err = c.repo.Burn.Create().
			SetExternalId(int64(typedParsedLog.ID)).
			SetSenderAddr(typedParsedLog.SenderAddr.String()).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
			SetTonMsg(tonMsg).
			Save(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to save burn: %v", err)
		}

	case *teleportcontract.ReinitLog:

		_, err = c.repo.Reinit.Create().
			SetExternalId(int64(typedParsedLog.ID)).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
			SetTonMsg(tonMsg).
			Save(c.ctx)
		if err != nil {
			log.Printf("[LogManager] failed to save reinit: %v", err)
		}

	default:
		log.Printf("[LogManager] unknown log type %T\n", typedParsedLog.GetLogID())
	}
}
