package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"k8s.io/client-go/util/workqueue"

	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/config"
	"github.com/rsquad/ton-teleport-btc-periphery/indexer/internal/log_listener"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/teleport_contract"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
)

type App struct {
	TonCenterV3Client           *ton.TonCenterV3Client
	TeleportContractLogListener loglistener.LogListenerInterface
	TeleportContractLogParser   *teleportcontract.LogParser
	TeleportLogsQueue           *workqueue.Typed[*cell.Cell]
}

func main() {
	log.SetFlags(0)

	app, err := initialize()
	if err != nil {
		log.Fatalf("[App] failed to initialize: %v", err)
	}

	go startTCPHealthCheck(":3000")

	if err := run(app); err != nil {
		log.Fatalf("[App] stopped with error: %v", err)
	}
}

func initialize() (*App, error) {
	log.Println("[App] initializing...")

	indexerConfig, err := utils.LoadConfig[config.IndexerConfig]()
	if err != nil {
		return nil, err
	}

	tonCenterV3Client, err := ton.NewTonCenterV3Client(
		indexerConfig.TonCenterV3Host,
		indexerConfig.TonCenterApiKey,
		"/",
		"https",
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create ton client: %w", err)
	}

	teleportContractAddr := address.MustParseAddr(indexerConfig.TeleportContractAddr)
	teleportLogsQueue := workqueue.NewTyped[*cell.Cell]()
	teleportContractLogListener, err := loglistener.NewLogListener(
		tonCenterV3Client,
		teleportContractAddr,
		teleportLogsQueue,
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create teleport contract log listener: %w", err)
	}

	teleportContractLogParser, err := teleportcontract.NewTeleportContractLogParser(
		teleportLogsQueue,
		[]*workqueue.Typed[*cell.Cell]{},
	)
	if err != nil {
		return nil, fmt.Errorf("[App] failed to create teleport contract log parser: %w", err)
	}

	log.Println("[App] initialized")

	return &App{
		TonCenterV3Client:           tonCenterV3Client,
		TeleportContractLogListener: teleportContractLogListener,
		TeleportContractLogParser:   teleportContractLogParser,
		TeleportLogsQueue:           teleportLogsQueue,
	}, nil
}

func run(app *App) error {
	log.Println("[App] running...")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.TeleportContractLogListener.StartListen()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.TeleportContractLogParser.StartParse()
	}()

	wg.Wait()

	log.Println("[App] shutdown complete")
	return nil
}

func startTCPHealthCheck(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("[App] failed to start healthcheck server: %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[App] failed to accept healthcheck ping: %v", err)
			continue
		}

		log.Println("[App] healthcheck pong")
		conn.Close()
	}
}
