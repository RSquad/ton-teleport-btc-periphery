ALTER TABLE metrics_data DROP COLUMN payload_hash;
ALTER TABLE metrics_data ALTER COLUMN payload TYPE jsonb USING payload::jsonb;
ALTER TABLE metrics_data DROP COLUMN update_at;
ALTER TABLE metrics_data DROP COLUMN create_at;
ALTER TABLE metrics_data ADD COLUMN create_at TIMESTAMPTZ NOT NULL DEFAULT NOW();