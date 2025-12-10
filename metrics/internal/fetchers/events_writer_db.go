package fetchers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/ton"
)

type EventsWriterDB struct {
	ch chan ton.EventInterface
	db *sql.DB
}

func NewEventsWriterDB(
	ch chan ton.EventInterface,
	db *sql.DB,
) (*EventsWriterDB, error) {
	// Create writer
	writer := EventsWriterDB{
		ch: ch,
		db: db,
	}

	// Prepare DB
	err := writer.PrepareDB()
	if err != nil {
		return nil, err
	}

	//
	return &writer, nil
}

func (writer *EventsWriterDB) PrepareDB() error {
	// Check if the table `events_data` exists
	_, err := writer.db.Exec(`CREATE TABLE IF NOT EXISTS events_data (
    	id        BIGSERIAL PRIMARY KEY,
    	create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    	event_id  BIGINT NOT NULL,
    	addr      TEXT NOT NULL,
			tx_hash   BYTEA NOT NULL,
	    tx_lt     BIGINT NOT NULL,
	    tx_utime  TIMESTAMPTZ NOT NULL,
      body      JSONB NOT NULL
		)`)
	if err != nil {
		return err
	}

	// Check index `events_tx_lt_idx`
	_, err = writer.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS events_event_id_tx_lt_idx ON events_data (event_id, tx_lt DESC, tx_hash)`)
	if err != nil {
		return err
	}

	return nil
}

func (writer *EventsWriterDB) Work(ctx context.Context) {
	defer logger.Log.Info().Msg("EventsWriterDB: stopped")
	logger.DefaultLogStartWork("EventsWriterDB: starting...")

	// Wait for Payload
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Writer DB received shutdown signal...")
			return
		case event, ok := <-writer.ch:
			if !ok {
				logger.Log.Warn().Msg("DKG Executor channel closed")
				return
			}

			err := writer.Write(event)
			if err != nil {
				logger.Log.Error().Msg(fmt.Sprintf("EventsWriterDB error: %v", err))
			}
		}
	}
}

func (writer *EventsWriterDB) Write(event ton.EventInterface) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	raw := event.GetRaw()

	_, err = writer.db.Exec(`
    INSERT INTO events_data (event_id, addr, tx_hash, tx_lt, tx_utime, body) VALUES($1, $2, $3, $4, $5, $6) ON CONFLICT (event_id, tx_lt, tx_hash) DO NOTHING`,
		event.GetEventID(),
		raw.Addr.StringRaw(),
		raw.TxHash,
		raw.TxLT,
		raw.TxUtime,
		string(jsonData),
	)

	return err
}
