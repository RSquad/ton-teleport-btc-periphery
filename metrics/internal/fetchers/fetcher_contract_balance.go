package fetchers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/utils"
	"github.com/rsquad/ton-teleport-btc-periphery/metrics/internal/config"
	"github.com/xssnick/tonutils-go/address"
)

type FetcherContractBalance struct {
	db        *sql.DB
	tonClient *tonclient.TonClient
	addr      *address.Address
	name      string
	period    int64 // Fetch period (in seconds)
	watchdog  *utils.Watchdog
}

func NewFetcherContractBalance(
	db *sql.DB,
	tonClient *tonclient.TonClient,
	cfg *config.ServicesConfig,
	addr *address.Address,
	name string,
	watchdog *utils.Watchdog,
) *FetcherContractBalance {
	return &FetcherContractBalance{
		db:        db,
		tonClient: tonClient,
		addr:      addr,
		name:      name,
		period:    int64(cfg.ContractBalancesFetchPeriod),
		watchdog:  watchdog,
	}
}

func (fetcher *FetcherContractBalance) Work(ctx context.Context, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	defer logger.Log.Info().Msgf("FetcherContractBalance: '%s' stopped", fetcher.name)
	logger.Log.Info().Msgf("FetcherContractBalance: '%s' starting...", fetcher.name)

	err = fetcher.PrepareDB()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(fetcher.period) * time.Second)
	defer ticker.Stop()

	// Setup watchdog
	fetcher.watchdog.Watch("FetcherContractBalance" + fetcher.name)
	defer fetcher.watchdog.Unwatch("FetcherContractBalance" + fetcher.name)

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("FetcherContractBalance received shutdown signal...")
			return
		case <-ticker.C:
			fetcher.Fetch()
			fetcher.watchdog.Heartbeat("FetcherContractBalance" + fetcher.name)
		}
	}
}

func (fetcher *FetcherContractBalance) Fetch() {
	balance, err := fetcher.GetBalance()
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractBalances").
			Msg("failed to GetBalances")

		return
	}

	err = fetcher.WriteDB(balance)
	if err != nil {
		logger.Log.Error().Err(err).
			Str("component", "FetcherContractBalances").
			Msg("failed to WriteDB")

		return
	}
}

func (fetcher *FetcherContractBalance) GetBalance() (int64, error) {
	account, err := fetcher.tonClient.FetchAcc(fetcher.addr, nil)
	if err != nil {
		return 0, err
	}

	if !account.IsActive {
		return 0, fmt.Errorf("FetcherContractBalances: account is not active %s, error: %v", utils.AddrToRawString(fetcher.addr), err)
	}

	balanceCoins, err := fetcher.tonClient.GetBalance(fetcher.addr)
	if err != nil {
		return 0, fmt.Errorf("FetcherContractBalances: can`t get balance for contract: %s, error: %v", utils.AddrToRawString(fetcher.addr), err)
	}

	return int64(balanceCoins.Nano().Uint64()), nil
}

func (fetcher *FetcherContractBalance) WriteDB(balance int64) error {
	_, err := fetcher.db.Exec(
		`WITH last_record AS (
      SELECT value AS balance
      FROM metrics_balances
      WHERE name = $1
      ORDER BY id DESC
      LIMIT 1
    )
    INSERT INTO metrics_balances (name, value)
    SELECT $1, $2
    WHERE NOT EXISTS (SELECT 1 FROM last_record) OR $2 != (SELECT balance FROM last_record)`,
		fetcher.name,
		balance,
	)

	return err
}

func (fetcher *FetcherContractBalance) PrepareDB() error {
	// Check if the table `metrics_balances` exists
	_, err := fetcher.db.Exec(`CREATE TABLE IF NOT EXISTS metrics_balances (
    	id BIGSERIAL PRIMARY KEY,
			create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    	name text,
    	value bigint
		)`)
	if err != nil {
		return err
	}

	// Check index `metrics_balances_name_id_idx`
	_, err = fetcher.db.Exec(`CREATE INDEX IF NOT EXISTS metrics_balances_name_idx ON metrics_balances (name, id DESC)`)
	if err != nil {
		return err
	}

	return nil
}
