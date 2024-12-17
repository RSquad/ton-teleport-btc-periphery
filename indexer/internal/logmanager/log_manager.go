package logmanager

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	ent "github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/pegin"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/ent/generated/tonmsg"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinatorcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/loglistener"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/pegoutcontract"
	teleportcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleportcontract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/toncenterv3"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type LogManager struct {
	ctx                            context.Context
	repo                           *ent.Client
	teleportContract               *teleportcontract.TeleportContract
	teleportContractLogParser      *teleportcontract.LogParser
	teleportContractLogListener    *loglistener.LogListener
	coordinatorContractLogParser   *coordinatorcontract.LogParser
	coordinatorContractLogListener *loglistener.LogListener
	pegoutContractCode             *cell.Cell
}

func New(
	ctx context.Context,
	repo *ent.Client,
	tonCenterV3Client *toncenterv3.Client,
	teleportContract *teleportcontract.TeleportContract,
	coordinatorContractAddr *address.Address,
) (
	*LogManager,
	error,
) {
	teleportContractStorage, err := teleportContract.GetStorage()
	if err != nil {
		return nil, err
	}

	pegoutContractCode := teleportContractStorage.PegoutContractCode

	logManager := &LogManager{
		ctx:                ctx,
		repo:               repo,
		teleportContract:   teleportContract,
		pegoutContractCode: pegoutContractCode,
	}

	teleportContractLogParser, err := teleportcontract.NewLogParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create teleport contract log parser: %w", err)
	}
	logManager.teleportContractLogParser = teleportContractLogParser

	teleportContractLogListener, err := loglistener.New(
		tonCenterV3Client,
		teleportContract.Addr,
		logManager.onLogReceived,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create teleport contract log listener: %w", err)
	}
	logManager.teleportContractLogListener = teleportContractLogListener

	coordinatorContractLogParser, err := coordinatorcontract.NewLogParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinator contract log parser: %w", err)
	}
	logManager.coordinatorContractLogParser = coordinatorContractLogParser

	coordinatorContractLogListener, err := loglistener.New(
		tonCenterV3Client,
		coordinatorContractAddr,
		logManager.onLogReceived,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create coordinator contract log listener: %w", err)
	}
	logManager.coordinatorContractLogListener = coordinatorContractLogListener

	return logManager, nil
}

func (c *LogManager) Run() {
	go c.teleportContractLogListener.StartListen()
	go c.coordinatorContractLogListener.StartListen()
}

func (c *LogManager) onLogReceived(logCell *cell.Cell, msgHash string, msgCreatedAt time.Time) {
	parsedLog, err := c.parseLog(logCell)
	if err != nil || parsedLog == nil {
		log.Printf("failed to parse log %v", err)
		return
	}

	exists, err := c.checkMsgExists(msgHash)
	if err != nil || exists {
		return
	}

	tonMsg, err := c.createTonMsg(msgHash, msgCreatedAt)
	if err != nil {
		return
	}

	err = c.saveLog(tonMsg, parsedLog)
	if err != nil {
		log.Printf("failed to save log %v: %v", parsedLog.GetLogID(), err)
	}
}

func (c *LogManager) parseLog(logCell *cell.Cell) (teleportcontract.LogInterface, error) {
	parsedLog, err := c.teleportContractLogParser.Parse(logCell)
	if parsedLog == nil {
		parsedLog, err = c.coordinatorContractLogParser.Parse(logCell)
	}
	return parsedLog, err
}

func (c *LogManager) saveLog(tonMsg *ent.TonMsg, parsedLog teleportcontract.LogInterface) error {
	switch typedParsedLog := parsedLog.(type) {
	case *teleportcontract.MintLog:
		return c.saveMint(tonMsg, typedParsedLog)
	case *teleportcontract.BurnLog:
		return c.saveBurn(tonMsg, typedParsedLog)
	case *teleportcontract.ReinitLog:
		return c.saveReinit(tonMsg, typedParsedLog)
	case *coordinatorcontract.DKGCompletedLog:
		return c.saveInternalKey(tonMsg, typedParsedLog)
	default:
		log.Printf("unknown log type %T\n", typedParsedLog.GetLogID())
		return nil
	}
}

func (c *LogManager) saveMint(tonMsg *ent.TonMsg, typedParsedLog *teleportcontract.MintLog) error {
	return c.saveTransaction(func(tx *ent.Tx) error {
		existingPegin, err := tx.Pegin.Query().
			Where(pegin.BitcoinTxIdEQ(typedParsedLog.BitcoinTxID.String())).
			Only(c.ctx)
		if err == nil {
			_, err = tx.Mint.UpdateOne(existingPegin.Edges.Mint).
				SetTonMsg(tonMsg).
				Save(c.ctx)
			return err
		}

		mint, err := tx.Mint.Create().
			SetAmount(typedParsedLog.Amount.String()).
			SetStatus("SUCCESS").
			SetTonMsg(tonMsg).
			Save(c.ctx)
		if err != nil {
			return err
		}
		_, err = tx.Pegin.Create().
			SetReceiverAddr(utils.AddrToRawString(typedParsedLog.ReceiverAddr)).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetMint(mint).
			Save(c.ctx)
		return err
	})
}

func (c *LogManager) saveBurn(tonMsg *ent.TonMsg, typedParsedLog *teleportcontract.BurnLog) error {
	return c.saveTransaction(func(tx *ent.Tx) error {
		pegout, err := c.savePegoutTx(tx, typedParsedLog)
		if err != nil {
			return err
		}
		_, err = tx.Burn.Create().
			SetExternalId(int64(typedParsedLog.ID)).
			SetSenderAddr(utils.AddrToRawString(typedParsedLog.SenderAddr)).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
			SetTonMsg(tonMsg).
			SetPegout(pegout).
			Save(c.ctx)
		return err
	})
}

func (c *LogManager) saveReinit(tonMsg *ent.TonMsg, typedParsedLog *teleportcontract.ReinitLog) error {
	return c.saveTransaction(func(tx *ent.Tx) error {
		pegout, err := c.savePegoutTx(tx, typedParsedLog)
		if err != nil {
			return err
		}
		_, err = tx.Reinit.Create().
			SetExternalId(int64(typedParsedLog.ID)).
			SetAmount(typedParsedLog.Amount.String()).
			SetBitcoinTxId(typedParsedLog.BitcoinTxID.String()).
			SetBitcoinScript(hex.EncodeToString(typedParsedLog.BitcoinScript)).
			SetTonMsg(tonMsg).
			SetPegout(pegout).
			Save(c.ctx)
		return err
	})
}

func (c *LogManager) saveInternalKey(tonMsg *ent.TonMsg, typedParsedLog *coordinatorcontract.DKGCompletedLog) error {
	_, err := c.repo.InternalKey.Create().
		SetCompletedAt(typedParsedLog.CompletedAt).
		SetKey(hex.EncodeToString(typedParsedLog.Key)).
		SetTonMsg(tonMsg).
		Save(c.ctx)
	return err
}

func (c *LogManager) saveTransaction(fn func(tx *ent.Tx) error) error {
	tx, err := c.repo.Tx(c.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()
	return fn(tx)
}

func (c *LogManager) savePegoutTx(tx *ent.Tx, parsedLog teleportcontract.LogWithPegoutInterface) (*ent.Pegout, error) {
	initData := &pegoutcontract.InitData{
		ID:                   uint32(parsedLog.GetID()),
		Amount:               parsedLog.GetAmount(),
		BitcoinScript:        parsedLog.GetBitcoinScript(),
		TeleportContractAddr: c.teleportContract.Addr,
	}

	pegoutContract, err := pegoutcontract.NewFromStateInit(&pegoutcontract.StateInit{
		Code:     c.pegoutContractCode,
		InitData: initData,
	}, c.teleportContract.TonClient, c.ctx)
	if err != nil {
		return nil, err
	}

	pegout, err := tx.Pegout.Create().
		SetExternalId(int64(parsedLog.GetID())).
		SetAddr(utils.AddrToRawString(pegoutContract.Addr)).
		Save(c.ctx)

	return pegout, err
}

func (c *LogManager) checkMsgExists(msgHash string) (bool, error) {
	exists, err := c.repo.TonMsg.Query().
		Where(tonmsg.Hash(msgHash)).
		Exist(c.ctx)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c *LogManager) createTonMsg(msgHash string, msgCreatedAt time.Time) (*ent.TonMsg, error) {
	tonMsg, err := c.repo.TonMsg.Create().
		SetHash(msgHash).
		SetCreatedAt(msgCreatedAt).
		Save(c.ctx)
	if err != nil {
		log.Printf("failed to save ton msg: %v", err)
		return nil, err
	}
	return tonMsg, nil
}
