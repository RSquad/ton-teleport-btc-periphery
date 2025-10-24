package fetchers

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
    	create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    	type_id INT,
    	payload JSONB
		)`)
	if err != nil {
		return err
	}

	// Check index `metrics_type_id_idx_desc`
	_, err = writer.db.Exec(`CREATE INDEX IF NOT EXISTS metrics_type_id_idx_desc ON metrics_data (type_id, id DESC)`)
	if err != nil {
		return err
	}

	// dkg_until_ts
	_, err = writer.db.Exec(`ALTER TABLE metrics_data ADD COLUMN IF NOT EXISTS dkg_until_ts timestamptz`)
	if err != nil {
		return err
	}

	//
	_, err = writer.db.Exec(`CREATE OR REPLACE FUNCTION safe_timestamptz(in_text text)
			RETURNS timestamptz
			LANGUAGE plpgsql STABLE PARALLEL SAFE AS $$
			BEGIN
				IF in_text IS NULL OR in_text = '' THEN
					RETURN NULL;
				END IF;

				RETURN in_text::timestamptz;
			EXCEPTION WHEN others THEN
				RETURN NULL;
			END;
			$$;`)
	if err != nil {
		return err
	}

	//
	_, err = writer.db.Exec(`CREATE OR REPLACE FUNCTION metrics_data_set_dkg_until_ts()
			RETURNS trigger
			LANGUAGE plpgsql AS $$
			BEGIN
				NEW.dkg_until_ts := safe_timestamptz(NEW.payload->>'Until');
				RETURN NEW;
			END;
			$$;`)
	if err != nil {
		return err
	}

	//
	_, err = writer.db.Exec(`DO $$
			BEGIN
				BEGIN
					EXECUTE $ddl$
						CREATE TRIGGER  metrics_data_dkg_until_ts_trg
						BEFORE INSERT OR UPDATE OF payload
						ON metrics_data
						FOR EACH ROW
						EXECUTE FUNCTION metrics_data_set_dkg_until_ts();
					$ddl$;
				EXCEPTION WHEN duplicate_object THEN
					-- trigger already exists: do nothing
				END;
			END$$`)
	if err != nil {
		return err
	}

	//
	_, err = writer.db.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS metrics_data_t1_dkg_until_id_desc_idx
  ON metrics_data (dkg_until_ts, id DESC)
  WHERE type_id = 1 AND dkg_until_ts IS NOT NULL`)
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
		`WITH last_record AS (
      SELECT md5(payload::text) AS payload_hash
      FROM metrics_data
      WHERE type_id = $1
      ORDER BY id DESC
      LIMIT 1
    )
    INSERT INTO metrics_data (type_id, payload)
    SELECT $1, $2::jsonb
    WHERE NOT EXISTS (SELECT 1 FROM last_record) OR md5(($2::jsonb)::text) != (SELECT payload_hash FROM last_record)`,
		payload.typeId,
		payload.payload,
	)

	return err
}
