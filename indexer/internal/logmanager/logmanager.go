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
	coordinatorcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/loglistener"
	teleportcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleport_contract"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type LogManager struct {
	ctx                            context.Context
	repo                           *ent.Client
	teleportContractLogParser      *teleportcontract.LogParser
	teleportContractLogListener    *loglistener.LogListener
	coordinatorContractLogParser   *coordinatorcontract.LogParser
	coordinatorContractLogListener *loglistener.LogListener
}

func New(
	ctx context.Context,
	repo *ent.Client,
	tonCenterV3Client *ton.TonCenterV3Client,
	teleportContractAddr *address.Address,
	coordinatorContractAddr *address.Address,
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

	teleportContractLogListener, err := loglistener.New(
		tonCenterV3Client,
		teleportContractAddr,
		logManager.onLogReceived,
	)
	if err != nil {
		return nil, fmt.Errorf("[LogManager] failed to create teleport contract log listener: %w", err)
	}
	logManager.teleportContractLogListener = teleportContractLogListener

	coordinatorContractLogParser, err := coordinatorcontract.NewLogParser()
	if err != nil {
		return nil, fmt.Errorf("[LogManager] failed to create coordinator contract log parser: %w", err)
	}
	logManager.coordinatorContractLogParser = coordinatorContractLogParser

	coordinatorContractLogListener, err := loglistener.New(
		tonCenterV3Client,
		coordinatorContractAddr,
		logManager.onLogReceived,
	)
	if err != nil {
		return nil, fmt.Errorf("[LogManager] failed to create coordinator contract log listener: %w", err)
	}
	logManager.coordinatorContractLogListener = coordinatorContractLogListener

	return logManager, nil
}

func (c *LogManager) Run() {
	go c.teleportContractLogListener.StartListen()
	go c.coordinatorContractLogListener.StartListen()
}

func (c *LogManager) onLogReceived(logCell *cell.Cell, msgHash string, msgCreatedAt time.Time) {
	parsedLog, err := c.teleportContractLogParser.Parse(logCell)
	if parsedLog == nil {
		parsedLog, err = c.coordinatorContractLogParser.Parse(logCell)
	}
	if err != nil {
		log.Printf("[LogManager] failed to parse log %v", err)
	}

	exists, err := c.checkMsgExists(msgHash)
	if err != nil || exists {
		return
	}

	tonMsg, err := c.createTonMsg(msgHash, msgCreatedAt)
	if err != nil {
		return
	}

	switch typedParsedLog := parsedLog.(type) {
	case *teleportcontract.MintLog:
		c.saveMint(tonMsg, typedParsedLog)
	case *teleportcontract.BurnLog:
		c.saveBurn(tonMsg, typedParsedLog)
	case *teleportcontract.ReinitLog:
		c.saveReinit(tonMsg, typedParsedLog)
	case *coordinatorcontract.DKGCompletedLog:
		c.saveInternalKey(tonMsg, typedParsedLog)
	default:
		log.Printf("%T", parsedLog)
		log.Printf("[LogManager] unknown log type %T\n", typedParsedLog.GetLogID())
	}
}

func (c *LogManager) saveMint(tonMsg *ent.TonMsg, typedParsedLog *teleportcontract.MintLog) {
	_, err := c.repo.Mint.Create().
		SetReceiverAddr(typedParsedLog.ReceiverAddr.String()).
		SetAmount(typedParsedLog.Amount.String()).
		SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
		SetTonMsg(tonMsg).
		Save(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to save mint: %v", err)
	}
}

func (c *LogManager) saveBurn(tonMsg *ent.TonMsg, typedParsedLog *teleportcontract.BurnLog) {
	_, err := c.repo.Burn.Create().
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
}

func (c *LogManager) saveReinit(tonMsg *ent.TonMsg, typedParsedLog *teleportcontract.ReinitLog) {
	_, err := c.repo.Reinit.Create().
		SetExternalId(int64(typedParsedLog.ID)).
		SetAmount(typedParsedLog.Amount.String()).
		SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
		SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
		SetTonMsg(tonMsg).
		Save(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to save reinit: %v", err)
	}
}

func (c *LogManager) saveInternalKey(tonMsg *ent.TonMsg, typedParsedLog *coordinatorcontract.DKGCompletedLog) {
	_, err := c.repo.InternalKey.Create().
		SetCompletedAt(typedParsedLog.CompletedAt).
		SetKey(hex.EncodeToString(typedParsedLog.Key)).
		SetTonMsg(tonMsg).
		Save(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to save internal key: %v", err)
	}
}

func (c *LogManager) checkMsgExists(msgHash string) (bool, error) {
	exists, err := c.repo.TonMsg.Query().
		Where(tonmsg.Hash(msgHash)).
		Exist(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to check ton msg existence: %v", err)
		return false, err
	}
	if exists {
		log.Printf("[LogManager] msg with hash %s already exists, skipping", msgHash)
	}
	return exists, nil
}

func (c *LogManager) createTonMsg(msgHash string, msgCreatedAt time.Time) (*ent.TonMsg, error) {
	tonMsg, err := c.repo.TonMsg.Create().
		SetHash(msgHash).
		SetCreatedAt(msgCreatedAt).
		Save(c.ctx)
	if err != nil {
		log.Printf("[LogManager] failed to save ton msg: %v", err)
		return nil, err
	}
	return tonMsg, nil
}
