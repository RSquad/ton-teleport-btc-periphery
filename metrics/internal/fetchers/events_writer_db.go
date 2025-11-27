package fetchers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type EventsWriterDB struct {
	ch chan PayloadDB
	db *sql.DB
}

func NewEventsWriterDB(
	ch chan PayloadDB,
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
    	id BIGSERIAL PRIMARY KEY,
    	create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    	type_id INT,
    	payload JSONB
		)`)
	if err != nil {
		return err
	}

	// Check index `events_type_id_idx_desc`
	_, err = writer.db.Exec(`CREATE INDEX IF NOT EXISTS events_type_id_idx_desc ON events_data (type_id, id DESC)`)
	if err != nil {
		return err
	}

	return nil
}

func (writer *EventsWriterDB) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("EventsWriterDB: stopped")
	logger.DefaultLogStartWork("EventsWriterDB: starting...")

	// Wait for Payload
	for {
		select {
		case <-ctx.Done():
			logger.Log.Info().Msg("Writer DB received shutdown signal...")
			return
		case payload, ok := <-writer.ch:
			if !ok {
				logger.Log.Warn().Msg("DKG Executor channel closed")
				return
			}

			err := writer.Write(payload)
			if err != nil {
				logger.Log.Error().Msg(fmt.Sprintf("EventsWriterDB: failed to retrieve CandidateBlockHashes, error: %v", err))
			}
		}
	}
}

func (writer *EventsWriterDB) Write(payload PayloadDB) error {
	_, err := writer.db.Exec(
		`WITH last_record AS (
      SELECT md5(payload::text) AS payload_hash
      FROM events_data
      WHERE type_id = $1
      ORDER BY id DESC
      LIMIT 1
    )
    INSERT INTO events_data (type_id, payload)
    SELECT $1, $2::jsonb
    WHERE NOT EXISTS (SELECT 1 FROM last_record) OR md5(($2::jsonb)::text) != (SELECT payload_hash FROM last_record)`,
		payload.typeId,
		payload.payload,
	)

	return err
}
