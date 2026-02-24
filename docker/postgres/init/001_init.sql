CREATE TABLE IF NOT EXISTS records (
    id BIGSERIAL PRIMARY KEY,
    data BYTEA NOT NULL,
    send_time TIMESTAMPTZ NOT NULL,
    rec_stat TEXT NOT NULL,
    send_chan TEXT NOT NULL,
    "from" TEXT NOT NULL DEFAULT '',
    "to" TEXT[] NOT NULL DEFAULT '{}',
    subject TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_records_send_time ON records (send_time);
CREATE INDEX IF NOT EXISTS idx_records_rec_stat ON records (rec_stat);
