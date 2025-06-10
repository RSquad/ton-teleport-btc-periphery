package metrics

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/coordinator"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton/tonclient"
)

type FetcherUnsignedPegouts struct {
	tonClient           *tonclient.TonClient
	db                  *sql.DB
	coordinatorContract coordinator.Coordinator
}

type PegoutTime struct {
	CreateAt time.Time
	UpdateAt time.Time
	result   ContractCoordinatorData
}

func NewFetcherUnsignedPegouts(tonClient *tonclient.TonClient, coordinator coordinator.Coordinator) *FetcherUnsignedPegouts {
	return &FetcherUnsignedPegouts{
		tonClient:           tonClient,
		coordinatorContract: coordinator,
	}
}

func (f *FetcherUnsignedPegouts) getPegoutTime() (PegoutTime, error) {
	rows, err := f.db.Query(
		`SELECT 
			create_at,
			update_at,
			jsonb_build_object(
				'contractCoordinator', (
					payload::json
				)
			) AS result
		FROM 
			metrics_data
		WHERE 
			type_id = 5
		ORDER BY id DESC
		LIMIT 1`,
	)
	if err != nil {
		return PegoutTime{}, err
	}

	defer rows.Close()

	var data PegoutTime
	if rows.Next() {
		err = rows.Scan(&data.CreateAt, &data.UpdateAt, &data.result)
		if err != nil {
			return PegoutTime{}, err
		}
	}

	return data, nil
}

func (fetcher *FetcherUnsignedPegouts) Work(ctx context.Context, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	defer logger.Log.Info().Msg("FetcherUnsignedPegouts: stopped")
	logger.DefaultLogStartWork("FetcherUnsignedPegouts: starting...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			unsignedPegouts, err := fetcher.coordinatorContract.GetUnsignedPegouts()
			if err != nil {
				return err
			}

			if unsignedPegouts == nil {
				logger.Log.Debug().Msg("FetcherUnsignedPegouts: Contract returns unsignedPegouts is null")
				return nil
			}

			if len(unsignedPegouts) == 0 {
				logger.Log.Debug().Msg("FetcherUnsignedPegouts: Contract returns unsignedPegouts is empty")
				return nil
			}

			pegoutTime, err := fetcher.getPegoutTime()
			if err != nil {
				return err
			}

			for _, value := range pegoutTime.result.UnsignedPegouts {
				if time.Now().After(pegoutTime.UpdateAt.Add(time.Minute * 20)) {
					unsignedPegoutDelayed.WithLabelValues(fmt.Sprint(value.ID)).Set(1)
				}
			}

			for _, value := range unsignedPegouts {
				if value.ExpiredAt.After(time.Now()) {
					unsignedPegoutRestart.WithLabelValues(fmt.Sprint(value.ID)).Set(1)
				}
			}

			unsignedPegoutsLen.WithLabelValues("Unsigned pegouts length").Set(float64(len(unsignedPegouts)))

			// TODO: reimplement with time.NewTicker
			time.Sleep(10 * time.Second)
		}
	}
}
