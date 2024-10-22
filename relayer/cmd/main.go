package main

import (
    "context"
    "encoding/hex"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/bitcoin"
    "github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
    "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/config"
    "github.com/rsquad/ton-teleport-btc-periphery/relayer/internal/relayer_factory"
)

type App struct {
    TonClient      *ton.Client
    BitcoinClient  *bitcoin.Client
    WalletContract *ton.WalletContract
    RelayerFactory *relayerfactory.RelayerFactory
}

func main() {
    app, err := initialize()
    if err != nil {
        log.Fatalf("Failed to initialize app: %v", err)
    }

    go startTCPHealthCheck(":3000")

    if err := run(app); err != nil {
        log.Fatalf("Application stopped with error: %v", err)
    }
}

func initialize() (*App, error) {
    log.Println("initializing app")

    config.LoadEnv()

    tonClient, err := ton.NewClient()
    if err != nil {
        return nil, fmt.Errorf("failed to create ton client: %v", err)
    }

    bitcoinClient, err := bitcoin.NewClient()
    if err != nil {
        return nil, fmt.Errorf("failed to create bitcoin client: %v", err)
    }

    walletV4SecretHex := os.Getenv("RELAYER_WALLET_V4_SECRET")
    walletV4Secret, err := hex.DecodeString(walletV4SecretHex)
    if err != nil {
        return nil, fmt.Errorf("failed to decode wallet secret: %v", err)
    }

    walletContract, err := ton.NewWalletContract(tonClient.API, walletV4Secret, context.Background())
    if err != nil {
        return nil, fmt.Errorf("failed to create wallet contract: %v", err)
    }

    relayerFactory := relayerfactory.NewRelayerFactory(bitcoinClient, tonClient)

    log.Println("app initialized")

    return &App{
        TonClient:      tonClient,
        BitcoinClient:  bitcoinClient,
        WalletContract: walletContract,
        RelayerFactory: relayerFactory,
    }, nil
}

func run(app *App) error {
    log.Println("running app")

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    if err := startRelayer(app, "block", 10*time.Second, ctx); err != nil {
        return fmt.Errorf("failed to start block relayer: %v", err)
    }

    time.Sleep(5 * time.Second)
    if err := startRelayer(app, "pegout", 20*time.Second, ctx); err != nil {
        return fmt.Errorf("failed to start pegout relayer: %v", err)
    }

    sig := <-sigCh
    log.Printf("received signal: %v. initiating shutdown...", sig)
    cancel()

    log.Println("shutdown complete")
    return nil
}

func startRelayer(app *App, relayerName string, interval time.Duration, ctx context.Context) error {
    relayer, err := app.RelayerFactory.CreateRelayer(relayerName, app.WalletContract)
    if err != nil {
        return fmt.Errorf("failed to create %v relayer: %v", relayerName, err)
    }

    ticker := time.NewTicker(interval)

    go func() {
        defer ticker.Stop()
        log.Printf("%v relayer started", relayerName)
        for {
            select {
            case <-ctx.Done():
                log.Printf("shutting down %v relayer", relayerName)
                return
            case <-ticker.C:
                if err := relayer.Relay(); err != nil {
                    log.Printf("failed to relay %v: %v", relayerName, err)
                }
            }
        }
    }()

    return nil
}

func startTCPHealthCheck(address string) {
    listener, err := net.Listen("tcp", address)
    if err != nil {
        log.Fatalf("failed to start tcp health check server: %v", err)
    }
    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("failed to accept connection: %v", err)
            continue
        }

        log.Println("health check received")
        conn.Close()
    }
}
