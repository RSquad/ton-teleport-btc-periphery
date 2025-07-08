package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/rsquad/ton-teleport-btc-periphery/lib/pkg/logger"
)

type WriterDB struct {
	ch chan PayloadDB
	db *sql.DB
}

func NewWriterDB(
	ch chan PayloadDB,
	db *sql.DB,
) (*WriterDB, error) {
	// Create writer
	writer := WriterDB{
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

func (writer *WriterDB) PrepareDB() error {
	// Check if the table `metrics_data` exists
	_, err := writer.db.Exec(`CREATE TABLE IF NOT EXISTS metrics_data (
    	id BIGSERIAL PRIMARY KEY,
    	create_at TIMESTAMPTZ NOT NULL,
			update_at TIMESTAMPTZ NOT NULL,
    	type_id INT,
    	payload TEXT,
			payload_hash TEXT GENERATED ALWAYS AS (md5(payload)) STORED,
			UNIQUE(type_id, payload_hash)
		);`)
	if err != nil {
		return err
	}

	return nil
}

func (writer *WriterDB) Work(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	defer logger.Log.Info().Msg("WriterDB: stopped")
	logger.DefaultLogStartWork("WriterDB: starting...")

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
				logger.Log.Error().Msg(fmt.Sprintf("WriterDB: failed to retrieve CandidateBlockHashes, error: %v", err))
			}
		}
	}
}

func (writer *WriterDB) Write(payload PayloadDB) error {
	_, err := writer.db.Exec(
		`INSERT INTO metrics_data (create_at, update_at, type_id, payload) VALUES 
		(
			TO_TIMESTAMP($1) AT TIME ZONE 'UTC',
			TO_TIMESTAMP($1) AT TIME ZONE 'UTC',
			$2,
			$3
		) ON CONFLICT ON CONSTRAINT "metrics_data_type_id_payload_hash_key" DO UPDATE SET update_at = NOW()`,
		payload.timestamp.Unix(),
		payload.typeId,
		payload.payload,
	)

	return err
}
