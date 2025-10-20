CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    linux_do_id TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    kyx_user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS claim_records (
    id BIGSERIAL PRIMARY KEY,
    linux_do_id TEXT NOT NULL,
    username TEXT NOT NULL,
    quota_added BIGINT NOT NULL,
    date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_claim_records_linux_do_id_date
    ON claim_records (linux_do_id, date);

CREATE TABLE IF NOT EXISTS donate_records (
    id BIGSERIAL PRIMARY KEY,
    linux_do_id TEXT NOT NULL,
    username TEXT NOT NULL,
    keys_count INT NOT NULL,
    total_quota_added BIGINT NOT NULL,
    push_status TEXT NOT NULL,
    push_message TEXT NOT NULL DEFAULT '',
    failed_keys JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_donate_records_linux_do_id
    ON donate_records (linux_do_id);

CREATE TABLE IF NOT EXISTS donated_keys (
    key_value TEXT PRIMARY KEY,
    linux_do_id TEXT NOT NULL,
    username TEXT NOT NULL,
    donate_record_id BIGINT REFERENCES donate_records(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_used BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS admin_config (
    id INT PRIMARY KEY,
    session TEXT NOT NULL DEFAULT '',
    new_api_user TEXT NOT NULL DEFAULT '1',
    claim_quota BIGINT NOT NULL DEFAULT 20000000,
    keys_api_url TEXT NOT NULL,
    keys_authorization TEXT NOT NULL DEFAULT '',
    group_id INT NOT NULL DEFAULT 26,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO admin_config (id, keys_api_url)
    VALUES (1, 'https://gpt-load.kyx03.de/api/keys/add-async')
ON CONFLICT (id) DO NOTHING;
