package logmanager

import (
	"fmt"
	"log"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	teleportcontract "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleport_contract"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type LogManager struct {
	teleportContractLogParser   *teleportcontract.LogParser
	teleportContractLogListener *teleportcontract.TeleportContractLogListener
}

func New(tonCenterV3Client *ton.TonCenterV3Client, teleportContractAddr *address.Address) (
	*LogManager,
	error,
) {
	logManager := &LogManager{}

	teleportContractLogParser, err := teleportcontract.NewTeleportContractLogParser()
	if err != nil {
		return nil, fmt.Errorf("[LogManager] failed to create teleport contract log parser: %w", err)
	}
	logManager.teleportContractLogParser = teleportContractLogParser

	teleportContractLogListener, err := teleportcontract.NewTeleportContractLogListener(
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

func (c *LogManager) onLogReceived(logCell *cell.Cell) {
	parsedLog, err := c.teleportContractLogParser.Parse(logCell)
	if err != nil {
		log.Printf("[LogManager] failed to parse log %v", err)
	}
	switch typedParsedLog := parsedLog.(type) {
	case *teleportcontract.MintLog:
		log.Printf("[LogManager] mint log: %v", typedParsedLog)
	case *teleportcontract.BurnLog:
		log.Printf("[LogManager] burn log: %v", typedParsedLog)
	case *teleportcontract.ReinitLog:
		log.Printf("[LogManager] reinit log: %v", typedParsedLog)
	default:
		log.Printf("[LogParser] unknown log type %T\n", typedParsedLog.GetLogID())
	}
}
